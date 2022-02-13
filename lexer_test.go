package lexer_test

import (
	"fmt"
	"strings"
	"testing"

	lexer "github.com/Acetolyne/commentlex"
)

//Ensures that each line gets checked by all the extensions available
//Prior issues with lines not being checked by extensions that were previously checked by prior lines
func TestCommentScanAllExtensions(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Init("tests/test.php")
	s.Mode = lexer.ScanComments
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", "")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "#@todo Comment 1// @todo Comment 2/* Multiline   Comment */# Comment 3//Comment 4/* Multiline   @todo   Comment 2 */"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("ch was not checked against all extensions")
	}
}

// func TestTemp(t *testing.T) {
// 	var s lexer.Scanner
// 	s.Init("tests/test.go")
// 	s.Mode = lexer.ScanComments
// 	tok := s.Scan()
// 	for tok != lexer.EOF {
// 		if tok == lexer.Comment {
// 			fmt.Println(s.TokenText())
// 		}
// 		tok = s.Scan()
// 	}
// 	fmt.Println(s.TokenText())

// }
