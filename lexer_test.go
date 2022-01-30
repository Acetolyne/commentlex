package lexer_test

import (
	"fmt"
	"testing"

	lexer "github.com/Acetolyne/commentlex"
)

func TestTemp(t *testing.T) {
	var s lexer.Scanner
	s.Init("tests/test.go")
	s.Mode = lexer.ScanComments
	_ = s.Scan()
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
