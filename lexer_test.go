package lexer_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	lexer "github.com/Acetolyne/commentlex"
)

func TestTemp(t *testing.T) {
	var s lexer.Scanner
	s.Init("tests/test.go")
	s.Mode = lexer.ScanComments
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			fmt.Println(s.TokenText())
		}
		tok = s.Scan()
	}
	fmt.Println(s.TokenText())

}

// func TestUTF8Filter(t *testing.T) {
// 	path := "tests/test1"
// 	file, err := os.Open(path)
// 	if err != nil {
// 		t.Log("ERROR: could not open file", err)
// 	}
// 	//t.Fatal("TEST")
// 	defer file.Close()
// 	var s lexer.Scanner
// 	//s.Error = func(*lexer.Scanner, string) {} // ignore errors
// 	s.Init(path)
// 	s.Mode = lexer.ScanComments
// 	tok := s.Scan()
// 	var line = "\n"
// 	//fmt.Println(path)
// 	for tok != lexer.EOF {
// 		t.Log(s.TokenText())
// 		if tok == lexer.Comment {
// 			line += strconv.Itoa(s.Position.Line) + ")" + s.TokenText()
// 		}
// 		tok = s.Scan()
// 	}
// 	//t.Log(s.src)
// }

//@todo add more tests!
// func TestSingleCommentShell(t *testing.T) {
// 	want := "\t#Shell comment 1\n\t#Shell comment 2\n\t#@todo Shell comment 3"
// 	path := "tests/test2.sh"
// 	var s lexer.Scanner
// 	s.Init(path)
// 	s.Mode = lexer.ScanComments
// 	checklines := func(s lexer.Scanner, path string) string {
// 		tok := s.Scan()
// 		var line string
// 		//fmt.Println(path)
// 		for tok != lexer.EOF {
// 			if tok == lexer.Comment {
// 				line += "\t" + s.TokenText()
// 			}
// 			tok = s.Scan()
// 		}

// 		return line
// 	}
// 	filelines := checklines(s, path)
// 	if filelines != want {
// 		t.Errorf("got: %q\nwant: %q", filelines, want)
// 	}
// }

// func TestSingleMatchingArg(t *testing.T) {
// 	want := "\t#@todo Shell comment 3"
// 	path := "tests/test.go"
// 	var s lexer.Scanner
// 	s.Init(path)
// 	s.Match = "@todo"
// 	s.Mode = lexer.ScanComments
// 	checklines := func(s lexer.Scanner, path string) string {
// 		tok := s.Scan()
// 		var line string
// 		//fmt.Println(path)
// 		for tok != lexer.EOF {
// 			if tok == lexer.Comment {
// 				line += "\t" + s.TokenText()
// 			}
// 			tok = s.Scan()
// 		}

// 		return line
// 	}
// 	filelines := checklines(s, path)
// 	if filelines != want {
// 		t.Errorf("got: %q\nwant: %q", filelines, want)
// 	}
// }

func TestScanner_ScanComment(t *testing.T) {
	type fields struct {
		src                   io.Reader
		singlePossible        bool
		multiPossible         bool
		Match                 string
		srcBuf                [bufLen + 1]byte
		srcPos                int
		srcEnd                int
		srcType               string
		srcBufOffset          int
		line                  int
		column                int
		lastLineLen           int
		lastCharLen           int
		CurSingleComment      string
		CurMultiStart         string
		CurMultiEnd           string
		CommentStatusSingle   map[int]string
		CommentStatusMulti    map[int]string
		CommentStatusMultiEnd map[int]string
		MultiExtNum           int
		tokBuf                bytes.Buffer
		tokPos                int
		tokEnd                int
		ch                    rune
		Error                 func(s *Scanner, msg string)
		ErrorCount            int
		Mode                  uint
		Whitespace            uint64
		IsIdentRune           func(ch rune, i int) bool
		Position              Position
	}
	type args struct {
		ch rune
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   rune
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				src:                   tt.fields.src,
				singlePossible:        tt.fields.singlePossible,
				multiPossible:         tt.fields.multiPossible,
				Match:                 tt.fields.Match,
				srcBuf:                tt.fields.srcBuf,
				srcPos:                tt.fields.srcPos,
				srcEnd:                tt.fields.srcEnd,
				srcType:               tt.fields.srcType,
				srcBufOffset:          tt.fields.srcBufOffset,
				line:                  tt.fields.line,
				column:                tt.fields.column,
				lastLineLen:           tt.fields.lastLineLen,
				lastCharLen:           tt.fields.lastCharLen,
				CurSingleComment:      tt.fields.CurSingleComment,
				CurMultiStart:         tt.fields.CurMultiStart,
				CurMultiEnd:           tt.fields.CurMultiEnd,
				CommentStatusSingle:   tt.fields.CommentStatusSingle,
				CommentStatusMulti:    tt.fields.CommentStatusMulti,
				CommentStatusMultiEnd: tt.fields.CommentStatusMultiEnd,
				MultiExtNum:           tt.fields.MultiExtNum,
				tokBuf:                tt.fields.tokBuf,
				tokPos:                tt.fields.tokPos,
				tokEnd:                tt.fields.tokEnd,
				ch:                    tt.fields.ch,
				Error:                 tt.fields.Error,
				ErrorCount:            tt.fields.ErrorCount,
				Mode:                  tt.fields.Mode,
				Whitespace:            tt.fields.Whitespace,
				IsIdentRune:           tt.fields.IsIdentRune,
				Position:              tt.fields.Position,
			}
			if got := s.ScanComment(tt.args.ch); got != tt.want {
				t.Errorf("Scanner.ScanComment() = %v, want %v", got, tt.want)
			}
		})
	}
}
