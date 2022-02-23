package lexer_test

import (
	"fmt"
	"strings"
	"testing"

	lexer "github.com/Acetolyne/commentlex"
)

//Ensures that each line gets checked by all the extensions available
//Prior issues with lines not being checked by extensions that were previously checked by prior lines
//@todo the wants should be global variables that are the same across all file types
func TestCommentScanAllExtensions(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Init("tests/test.php")
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

func TestCanUseMatchArgument(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Match = "@todo"
	s.Init("tests/test.php")
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", "")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "#@todo Comment 1// @todo Comment 2/* Multiline   @todo   Comment 2 */"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("s.Match not working as expected")
	}
}

func TestUseFileExtensions(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Match = "@todo"
	s.Init("tests/test.lua")
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", " ")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "--@todo singleline todo comment--[[Multiline Lua Comment with @todo in it --]]--@todo another inline comment"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("unable to use file extensions")
	}
}

func TestMultipleCommentTypesPerFile(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Match = "@todo"
	s.Init("tests/test.html")
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", " ")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "<!-- @todo some comment -->//@todo some javascript comment in html<!-- @todo multiline comment in  html -->"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("unable to detect multiple comment types in single file")
	}
}

func TestDefaultFileExtension(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Match = "@todo"
	s.Init("tests/test")
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", " ")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "//@todo Single Comment/* Multiline @todo some test Comment */"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("unable to detect multiple comment types in single file")
	}
}

func TestRubyCommentsWork(t *testing.T) {
	res := ""
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	s.Match = "@todo"
	s.Init("tests/test.rb")
	tok := s.Scan()
	for tok != lexer.EOF {
		if tok == lexer.Comment {
			line := strings.ReplaceAll(s.TokenText(), "\n", " ")
			res += strings.ReplaceAll(line, "\t", "")
		}
		tok = s.Scan()
	}

	want := "#@todo comment 1=begin @todo multiline comment 2 =end#@todo comment 3"
	if res != want {
		fmt.Println("got", res, "want", want)
		t.Fatalf("unable to use ruby comments")
	}
}
