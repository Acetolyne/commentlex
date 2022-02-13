package lexer_test

import (
	"fmt"
	"testing"

	lexer "github.com/Acetolyne/commentlex"
)

//Ensures that each line gets checked by all the extensions available
//Prior issues with lines not being checked by extensions that were previously checked by prior lines
func TestCommentScanAllExtensions(t *testing.T) {
	var s lexer.Scanner
	s.Init("tests/test.php")
	s.Mode = lexer.ScanComments
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			fmt.Println(s.TokenText())
		}
		tok = s.Scan()
	}

	// if total != 10 {
	// 	t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	// }
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
