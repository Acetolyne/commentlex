package main

import (
	"fmt"

	lexer "github.com/Acetolyne/commentlex"
)

type CommentValues struct {
	ext         []string
	startSingle string
	startMulti  string
	endMulti    string
}

func main() {
	var s lexer.Scanner
	s.Mode = lexer.ScanComments
	//s.Init("tests/test.php")
	allext := s.GetExtensions()
	//@todo instead of printing copy these values to the end of the README.md file
	for l := range allext {
		//fmt.Println(allext[l])
		//fmt.Println(t)
		curext := allext[l]
		fmt.Println(curext)
		// for filetype := range curext.ext {
		// 	fmt.Println(filetype)
		// }
	}
}
