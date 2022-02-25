package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	var buffer string
	s.Mode = lexer.ScanComments
	allext := s.GetExtensions()

	file, err := os.Open("../../README.md")
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		buffer += scanner.Text() + "\n"
		if strings.Contains(buffer, "##### Supported Filetypes") {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	//@todo instead of printing copy these values to the end of the README.md file
	for l := range allext {
		//fmt.Println(allext[l])
		//fmt.Println(t)
		curext := allext[l]
		buffer += curext + "\n"

		//@todo open readme file for reading
		//@todo read the file line by line until we find the line starting with "##### Supported Filetypes"
		//@todo on the next line start writing to file with "\n" plus the current file extension
		// for filetype := range curext.ext {
		// 	fmt.Println(filetype)
		// }
	}
	file.Close()
	file, err = os.OpenFile("../../README.md", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}
	err = file.Truncate(0)
	if err != nil {
		fmt.Println(err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
	}
	_, err = fmt.Fprintf(file, "%s", buffer)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(buffer)
}
