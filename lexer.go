// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scanner provides a scanner and tokenizer for UTF-8-encoded text.
// It takes an io.Reader providing the source, which then can be tokenized
// through repeated calls to the Scan function. For compatibility with
// existing tools, the NUL character is not allowed. If the first character
// in the source is a UTF-8 encoded byte order mark (BOM), it is discarded.
//
// By default, a Scanner skips white space and Go comments and recognizes all
// literals as defined by the Go language specification. It may be
// customized to recognize only a subset of those literals and to recognize
// different identifier and white space characters.
package lexer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

var SubCheck string

// Position is a value that represents a source position.
// A position is valid if Line > 0.
type Position struct {
	Filename string // filename, if any
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

type CommentValues struct {
	ext         []string
	startSingle string
	startMulti  string
	endMulti    string
}

//Initialize comment characters based on file extension
//This may be more than one type per filetype as html can have javascript comments in them as well and there may be other filetypes that have multiple languages in them
var Extensions = []CommentValues{
	{
		ext:         []string{".go", ".py", ".js", ".rs", ".html", ".gohtml", ".php"},
		startSingle: "//",
		startMulti:  "/*",
		endMulti:    "*/",
	},
	{
		ext:         []string{".sh", ".php"},
		startSingle: "#",
		startMulti:  "",
		endMulti:    "",
	},
	{
		ext:         []string{".html", ".gohtml"},
		startSingle: "",
		startMulti:  "<!--",
		endMulti:    "-->",
	},
	{
		ext:         []string{".lua"},
		startSingle: "--",
		startMulti:  "--[[",
		endMulti:    "--]]",
	},
}

// IsValid reports whether the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

func (pos Position) String() string {
	s := pos.Filename
	if s == "" {
		s = "<input>"
	}
	if pos.IsValid() {
		s += fmt.Sprintf(":%d:%d", pos.Line, pos.Column)
	}
	return s
}

// Predefined mode bits to control recognition of tokens. For instance,
// to configure a Scanner such that it only recognizes (Go) identifiers,
// integers, and skips comments, set the Scanner's Mode field to:
//
//	ScanIdents | ScanInts | SkipComments
//
// With the exceptions of comments, which are skipped if SkipComments is
// set, unrecognized tokens are not ignored. Instead, the scanner simply
// returns the respective individual characters (or possibly sub-tokens).
// For instance, if the mode is ScanIdents (not ScanStrings), the string
// "foo" is scanned as the token sequence '"' Ident '"'.
//
// Use GoTokens to configure the Scanner such that it accepts all Go
// literal tokens including Go identifiers. Comments will be skipped.
//
const (
	ScanIdents     = 1 << -Ident
	ScanInts       = 1 << -Int
	ScanFloats     = 1 << -Float // includes Ints and hexadecimal floats
	ScanChars      = 1 << -Char
	ScanStrings    = 1 << -String
	ScanRawStrings = 1 << -RawString
	ScanComments   = 1 << -Comment
	SkipComments   = 1 << -skipComment // if set with ScanComments, comments become white space
	GoTokens       = ScanIdents | ScanFloats | ScanChars | ScanStrings | ScanRawStrings | ScanComments | SkipComments
)

// The result of Scan is one of these tokens or a Unicode character.
const (
	EOF = -(iota + 1)
	Ident
	Int
	Float
	Char
	String
	RawString
	Comment

	// internal use only
	skipComment
)

var tokenString = map[rune]string{
	EOF:       "EOF",
	Ident:     "Ident",
	Int:       "Int",
	Float:     "Float",
	Char:      "Char",
	String:    "String",
	RawString: "RawString",
	Comment:   "Comment",
}

// TokenString returns a printable string for a token or Unicode character.
func TokenString(tok rune) string {
	if s, found := tokenString[tok]; found {
		return s
	}
	return fmt.Sprintf("%q", string(tok))
}

// GoWhitespace is the default value for the Scanner's Whitespace field.
// Its value selects Go's white space characters.
const GoWhitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

const bufLen = 1024 // at least utf8.UTFMax

// A Scanner implements reading of Unicode characters and tokens from an io.Reader.
type Scanner struct {
	// Input
	src io.Reader

	//Active Possible Comment Types
	singlePossible bool
	multiPossible  bool

	//Additional Characters to match after comment characters, this is a way to further filter the comments
	//This string must be directly after the comment characters for a single line comment or anywhere in a multiline comment
	Match string

	// Source buffer
	srcBuf  [bufLen + 1]byte // +1 for sentinel for common case of s.next()
	srcPos  int              // reading position (srcBuf index)
	srcEnd  int              // source end (srcBuf index)
	srcType string           // file extension for choosing the comment characters

	// Source position
	srcBufOffset int // byte offset of srcBuf[0] in source
	line         int // line count
	column       int // character count
	lastLineLen  int // length of last line in characters (for correct column reporting)
	lastCharLen  int // length of last character in bytes

	// Comment characters to search for based on file type
	CurSingleComment      string
	CurMultiStart         string
	CurMultiEnd           string
	CommentStatusSingle   map[int]string
	CommentStatusMulti    map[int]string
	CommentStatusMultiEnd map[int]string
	MultiExtNum           int

	// Token text buffer
	// Typically, token text is stored completely in srcBuf, but in general
	// the token text's head may be buffered in tokBuf while the token text's
	// tail is stored in srcBuf.
	tokBuf bytes.Buffer // token text head that is not in srcBuf anymore
	tokPos int          // token text tail position (srcBuf index); valid if >= 0
	tokEnd int          // token text tail end (srcBuf index)

	// One character look-ahead
	ch rune // character before current srcPos

	// Error is called for each error encountered. If no Error
	// function is set, the error is reported to os.Stderr.
	Error func(s *Scanner, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int

	// The Mode field controls which tokens are recognized. For instance,
	// to recognize Ints, set the ScanInts bit in Mode. The field may be
	// changed at any time.
	Mode uint

	// The Whitespace field controls which characters are recognized
	// as white space. To recognize a character ch <= ' ' as white space,
	// set the ch'th bit in Whitespace (the Scanner's behavior is undefined
	//for values ch > ' '). The field may be changed at any time.
	Whitespace uint64

	// IsIdentRune is a predicate controlling the characters accepted
	// as the ith rune in an identifier. The set of valid characters
	// must not intersect with the set of white space characters.
	// If no IsIdentRune function is set, regular Go identifiers are
	// accepted instead. The field may be changed at any time.
	IsIdentRune func(ch rune, i int) bool

	// Start position of most recently scanned token; set by Scan.
	// Calling Init or Next invalidates the position (Line == 0).
	// The Filename field is always left untouched by the Scanner.
	// If an error is reported (via Error) and Position is invalid,
	// the scanner is not inside a token. Call Pos to obtain an error
	// position in that case, or to obtain the position immediately
	// after the most recently scanned token.
	Position
}

// Init initializes a Scanner with a new source and returns s.
// Error is set to nil, ErrorCount is set to 0, Mode is set to GoTokens,
// and Whitespace is set to GoWhitespace.
func (s *Scanner) Init(file string) *Scanner {

	// All comment types that are possible when we first start scanning
	s.singlePossible = true
	s.multiPossible = true
	// for v := range Extensions {
	// 	s.CommentStatusSingle[v] = ""
	// }
	if s.CommentStatusSingle == nil {
		s.CommentStatusSingle = make(map[int]string)
	}
	if s.CommentStatusMulti == nil {
		s.CommentStatusMulti = make(map[int]string)
	}
	if s.CommentStatusMultiEnd == nil {
		s.CommentStatusMultiEnd = make(map[int]string)
	}

	// Get the filetype so we can set the comment characters for this scan
	src, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	s.src = src

	s.srcType = filepath.Ext(file)

	//@todo since we can only parse UTF-8 encoded files lets check if its valid UTF-8 before we do any operations on it

	// initialize source buffer
	// (the first call to next() will fill it by calling src.Read)
	s.srcBuf[0] = utf8.RuneSelf // sentinel
	s.srcPos = 0
	s.srcEnd = 0

	// initialize source position
	s.srcBufOffset = 0
	s.line = 1
	s.column = 0
	s.lastLineLen = 0
	s.lastCharLen = 0

	// initialize token text buffer
	// (required for first call to next()).
	s.tokPos = -1

	// initialize one character look-ahead
	s.ch = -2 // no char read yet, not EOF

	// initialize public fields
	s.Error = nil
	s.ErrorCount = 0
	s.Mode = GoTokens
	s.Whitespace = GoWhitespace
	s.Line = 0 // invalidate token position

	return s
}

// next reads and returns the next Unicode character. It is designed such
// that only a minimal amount of work needs to be done in the common ASCII
// case (one test to check for both ASCII and end-of-buffer, and one test
// to check for newlines).
func (s *Scanner) next() rune {
	ch, width := rune(s.srcBuf[s.srcPos]), 1

	if ch >= utf8.RuneSelf {
		// uncommon case: not ASCII or not enough bytes
		for s.srcPos+utf8.UTFMax > s.srcEnd && !utf8.FullRune(s.srcBuf[s.srcPos:s.srcEnd]) {
			// not enough bytes: read some more, but first
			// save away token text if any
			if s.tokPos >= 0 {
				s.tokBuf.Write(s.srcBuf[s.tokPos:s.srcPos])
				s.tokPos = 0
				// s.tokEnd is set by Scan()
			}
			// move unread bytes to beginning of buffer
			copy(s.srcBuf[0:], s.srcBuf[s.srcPos:s.srcEnd])
			s.srcBufOffset += s.srcPos
			// read more bytes
			// (an io.Reader must return io.EOF when it reaches
			// the end of what it is reading - simply returning
			// n == 0 will make this loop retry forever; but the
			// error is in the reader implementation in that case)
			i := s.srcEnd - s.srcPos
			n, err := s.src.Read(s.srcBuf[i:bufLen])
			s.srcPos = 0
			s.srcEnd = i + n
			s.srcBuf[s.srcEnd] = utf8.RuneSelf // sentinel
			if err != nil {
				if err != io.EOF {
					s.error(err.Error())
				}
				if s.srcEnd == 0 {
					if s.lastCharLen > 0 {
						// previous character was not EOF
						s.column++
					}
					s.lastCharLen = 0
					return EOF
				}
				// If err == EOF, we won't be getting more
				// bytes; break to avoid infinite loop. If
				// err is something else, we don't know if
				// we can get more bytes; thus also break.
				break
			}
		}
		// at least one byte
		ch = rune(s.srcBuf[s.srcPos])
		if ch >= utf8.RuneSelf {
			// uncommon case: not ASCII
			ch, width = utf8.DecodeRune(s.srcBuf[s.srcPos:s.srcEnd])
			if ch == utf8.RuneError && width == 1 {
				// advance for correct error position
				s.srcPos += width
				s.lastCharLen = width
				s.column++
				s.error("invalid UTF-8 encoding")
				return ch
			}
		}
	}

	// advance
	s.srcPos += width
	s.lastCharLen = width
	s.column++

	// special situations
	switch ch {
	case 0:
		// for compatibility with other tools
		s.error("invalid character NUL")
	case '\n':
		s.line++
		s.lastLineLen = s.column
		s.column = 0
	}

	return ch
}

// Next reads and returns the next Unicode character.
// It returns EOF at the end of the source. It reports
// a read error by calling s.Error, if not nil; otherwise
// it prints an error message to os.Stderr. Next does not
// update the Scanner's Position field; use Pos() to
// get the current position.
func (s *Scanner) Next() rune {
	s.tokPos = -1 // don't collect token text
	s.Line = 0    // invalidate token position
	ch := s.Peek()
	if ch != EOF {
		s.ch = s.next()
	}
	return ch
}

// Peek returns the next Unicode character in the source without advancing
// the scanner. It returns EOF if the scanner's position is at the last
// character of the source.
func (s *Scanner) Peek() rune {
	if s.ch == -2 {
		// this code is only run for the very first character
		s.ch = s.next()
		if s.ch == '\uFEFF' {
			s.ch = s.next() // ignore BOM
		}
	}
	return s.ch
}

func (s *Scanner) error(msg string) {
	s.tokEnd = s.srcPos - s.lastCharLen // make sure token text is terminated
	s.ErrorCount++
	if s.Error != nil {
		s.Error(s, msg)
		return
	}
	pos := s.Position
	if !pos.IsValid() {
		pos = s.Pos()
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", pos, msg)
}

func (s *Scanner) errorf(format string, args ...interface{}) {
	s.error(fmt.Sprintf(format, args...))
}

func (s *Scanner) isIdentRune(ch rune, i int) bool {
	if s.IsIdentRune != nil {
		return s.IsIdentRune(ch, i)
	}
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
}

func (s *Scanner) scanIdentifier() rune {
	// we know the zero'th rune is OK; start scanning at the next one
	ch := s.next()
	for i := 1; s.isIdentRune(ch, i); i++ {
		ch = s.next()
	}
	return ch
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

// digits accepts the sequence { digit | '_' } starting with ch0.
// If base <= 10, digits accepts any decimal digit but records
// the first invalid digit >= base in *invalid if *invalid == 0.
// digits returns the first rune that is not part of the sequence
// anymore, and a bitset describing whether the sequence contained
// digits (bit 0 is set), or separators '_' (bit 1 is set).
func (s *Scanner) digits(ch0 rune, base int, invalid *rune) (ch rune, digsep int) {
	ch = ch0
	if base <= 10 {
		max := rune('0' + base)
		for isDecimal(ch) || ch == '_' {
			ds := 1
			if ch == '_' {
				ds = 2
			} else if ch >= max && *invalid == 0 {
				*invalid = ch
			}
			digsep |= ds
			ch = s.next()
		}
	} else {
		for isHex(ch) || ch == '_' {
			ds := 1
			if ch == '_' {
				ds = 2
			}
			digsep |= ds
			ch = s.next()
		}
	}
	return
}

func litname(prefix rune) string {
	switch prefix {
	default:
		return "decimal literal"
	case 'x':
		return "hexadecimal literal"
	case 'o', '0':
		return "octal literal"
	case 'b':
		return "binary literal"
	}
}

// invalidSep returns the index of the first invalid separator in x, or -1.
func invalidSep(x string) int {
	x1 := ' ' // prefix char, we only care if it's 'x'
	d := '.'  // digit, one of '_', '0' (a digit), or '.' (anything else)
	i := 0

	// a prefix counts as a digit
	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		if x1 == 'x' || x1 == 'o' || x1 == 'b' {
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}

	return -1
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

//scanComment scans current line or lines and returns if it is a comment or not
func (s *Scanner) scanComment(ch rune) rune {
	// isSingle := false
	// isMulti := false
	// //multicheck := ""
	// for ch != '\n' && ch >= 0 {
	// 	for v := range Extensions {
	// 		curext := Extensions[v].ext
	// 		for ext := range curext {
	// 			if s.srcType == curext[ext] {
	// 				curlensingle := len(s.CommentStatusSingle[v])
	// 				//curlenmulti := len(s.CommentStatusMulti[v])
	// 				if Extensions[v].startSingle != "" {
	// 					if curlensingle < len(string(Extensions[v].startSingle)) {
	// 						if string(ch) == string(Extensions[v].startSingle[curlensingle]) { //If this character matches the current character in the extension then append it else clear it because characters are not consecutive
	// 							s.CommentStatusSingle[v] += string(ch)
	// 							if len(s.CommentStatusSingle[v]) == len(Extensions[v].startSingle) {
	// 								isSingle = true
	// 							}
	// 						} else {
	// 							s.CommentStatusSingle[v] = ""

	// 						}
	// 					}
	// 				}
	// 				// if Extensions[v].startMulti != "" {
	// 				// 	multicheck += string(ch)
	// 				// 	if curlenmulti < len(string(Extensions[v].startMulti)) {
	// 				// 		if string(ch) == string(Extensions[v].startMulti[curlenmulti]) { //If this character matches the current character in the extension then append it else clear it because characters are not consecutive
	// 				// 			s.CommentStatusMulti[v] += string(ch)
	// 				// 			if len(s.CommentStatusMulti[v]) == len(Extensions[v].startMulti) {
	// 				// 				isMulti = true
	// 				// 				s.MultiExtNum = v
	// 				// 			}
	// 				// 		} else {
	// 				// 			s.CommentStatusMulti[v] = ""

	// 				// 		}
	// 				// 	}
	// 				// }
	// 			}
	// 		}
	// 	}
	// 	ch = s.next()
	// }
	// //Always check multi first because in some languages multi starts with the same characters as single line comments (Lua)
	// //@todo fix issue where anything after multiline comment is not returned as a comment
	// fmt.Println("isMulti: ", isMulti, "isSingle:", isSingle)
	// for v := range Extensions {
	// 	if isMulti == true {
	// 		if s.CommentStatusMulti[v] == Extensions[v].startMulti {
	// 			// 	for ch != EOF {
	// 			//  multicheck += string(ch)
	// 			// 		curlenmultiend := len(s.CommentStatusMultiEnd[s.MultiExtNum])
	// 			// 		if Extensions[s.MultiExtNum].endMulti != "" {
	// 			// 			fmt.Println(Extensions[s.MultiExtNum].endMulti)
	// 			// 			if curlenmultiend < len(string(Extensions[s.MultiExtNum].endMulti)) {
	// 			// 				if string(ch) == string(Extensions[s.MultiExtNum].endMulti[curlenmultiend]) { //If this character matches the current character in the extension then append it else clear it because characters are not consecutive
	// 			// 					fmt.Println("adding ch")
	// 			// 					s.CommentStatusMultiEnd[s.MultiExtNum] += string(ch)
	// 			// 				} else {
	// 			// 					s.CommentStatusMultiEnd[s.MultiExtNum] = ""

	// 			// 				}
	// 			// 			}
	// 			// 			if curlenmultiend == len(string(Extensions[s.MultiExtNum].endMulti)) {
	// 			// 				fmt.Println("SM:", s.Match)
	// 			// 				//@todo below is temp since flowcat passes this value in by default for testing only should check if not null
	// 			// 				if s.Match != "" {
	// 			// 					fmt.Println("Trying to match")
	// 			// 					return Comment
	// 			// 					// if strings.Contains(multicheck, s.Match) {
	// 			// 					// 	return Comment
	// 			// 					// } else {
	// 			// 					// 	return ch
	// 			// 					// }
	// 			// 				} else {
	// 			// 					return Comment
	// 			// 				}
	// 			// 			}
	// 			// 		}
	// 			// 		fmt.Println("NEXT")
	// 			// 		ch = s.next()
	// 			// 	}
	// 		}
	// 		s.CommentStatusMulti[v] = ""
	// 		s.CommentStatusMultiEnd[v] = ""
	// 	}
	// }
	// for v := range Extensions {
	// 	if isSingle == true {
	// 		s.CommentStatusSingle[v] = ""
	// 		return Comment
	// 	}
	// }
	// return ch
	isSingle := false
	for ch >= 0 {
		for ch != EOF {
			if ch != '\n' {
				for v := range Extensions {
					if Extensions[v].startSingle != "" {
						fmt.Println(v, ")", string(ch), string(Extensions[v].startSingle[0]))
						if string(ch) == string(Extensions[v].startSingle[0]) {
							fmt.Println("setting true")
							isSingle = true
						}
					}
				}
			} else {
				fmt.Println("EOL", isSingle)
				if isSingle == true {
					isSingle = false
					return Comment
				}
				ch = s.next()
			}
			ch = s.next()
		}
		ch = s.next()
	}

	return ch
}

// Scan reads the next token or Unicode character from source and returns it.
// It only recognizes tokens t for which the respective Mode bit (1<<-t) is set.
// It returns EOF at the end of the source. It reports scanner errors (read and
// token errors) by calling s.Error, if not nil; otherwise it prints an error
// message to os.Stderr.

func (s *Scanner) Scan() rune {
	//go to the first character
	ch := s.next()

	// reset token text position
	s.tokPos = -1
	s.Line = 0

	// skip white space
	for s.Whitespace&(1<<uint(ch)) != 0 {
		ch = s.next()
	}

	// start collecting token text
	s.tokBuf.Reset()
	s.tokPos = s.srcPos - s.lastCharLen

	// set token position
	// (this is a slightly optimized version of the code in Pos())
	s.Offset = s.srcBufOffset + s.tokPos
	if s.column > 0 {
		// common case: last character was not a '\n'
		s.Line = s.line
		s.Column = s.column
	} else {
		// last character was a '\n'
		// (we cannot be at the beginning of the source
		// since we have called next() at least once)
		s.Line = s.line - 1
		s.Column = s.lastLineLen
	}

	// determine token value
	tok := ch
	switch ch {
	case EOF:
		break
	default:
		//@todo add more file types and comment characters
		tok := s.scanComment(ch)
		s.tokEnd = s.srcPos - s.lastCharLen
		s.ch = ch
		return tok
	}

	// end of token text
	s.tokEnd = s.srcPos - s.lastCharLen

	s.ch = ch
	return tok
}

// Pos returns the position of the character immediately after
// the character or token returned by the last call to Next or Scan.
// Use the Scanner's Position field for the start position of the most
// recently scanned token.
func (s *Scanner) Pos() (pos Position) {
	pos.Filename = s.Filename
	pos.Offset = s.srcBufOffset + s.srcPos - s.lastCharLen
	switch {
	case s.column > 0:
		// common case: last character was not a '\n'
		pos.Line = s.line
		pos.Column = s.column
	case s.lastLineLen > 0:
		// last character was a '\n'
		pos.Line = s.line - 1
		pos.Column = s.lastLineLen
	default:
		// at the beginning of the source
		pos.Line = 1
		pos.Column = 1
	}
	return
}

// TokenText returns the string corresponding to the most recently scanned token.
// Valid after calling Scan and in calls of Scanner.Error.
func (s *Scanner) TokenText() string {
	if s.tokPos < 0 {
		// no token text
		return ""
	}

	if s.tokEnd < s.tokPos {
		// if EOF was reached, s.tokEnd is set to -1 (s.srcPos == 0)
		s.tokEnd = s.tokPos
	}
	// s.tokEnd >= s.tokPos

	if s.tokBuf.Len() == 0 {
		// common case: the entire token text is still in srcBuf
		return string(s.srcBuf[s.tokPos:s.tokEnd])
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBuf[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}
