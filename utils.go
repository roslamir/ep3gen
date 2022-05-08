// Copyright (C) 2022, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// List of utility functions

package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	usage = `EP3Gen is a tool for generating EPUB3 e-books.

Usage: ep3gen [bookdir]

bookdir specifies the directory under ./data/source/ which contains the source artifacts used to create the EPUB3 e-book.`
)

// checkArgs checks the input arguments and acts accordingly.
func checkArgs() {
	// Show usage information if no arguments are given
	if len(os.Args) != 2 {
		fmt.Println(usage)
		os.Exit(1)
	}
	bookName = os.Args[1]
}

// nextLine returns the next source line.
func nextLine() string {
	lineIndex++
	if lineIndex == len(lines) {
		fmt.Println("ERROR: Unexpected end of file")
		os.Exit(1)
	}
	return lines[lineIndex]
}

// extractMetaData extracts the metadata 'name' and 'content' from the current line.
// Stores it into the 'attributes' map.
func extractMetaData() {
	index := strings.Index(currLine, "name=")
	if index == -1 {
		return
	}
	name := currLine[index+len("name=")+1:] // skip past 'name="'
	index = strings.Index(name, "\"")
	if index == -1 {
		fmt.Println("ERROR: Invalid 'meta' HTML line: ", currLine)
		os.Exit(1)
	}
	name = name[:index]

	index = strings.Index(currLine, "content=")
	if index == -1 {
		fmt.Println("ERROR: Invalid 'meta' HTML line: ", currLine)
		os.Exit(1)
	}
	content := currLine[index+len("content=")+1:] // skip past 'content="'
	index = strings.Index(content, "\"")
	if index == -1 {
		fmt.Println("ERROR: Invalid 'meta' HTML line: ", currLine)
		os.Exit(1)
	}
	content = content[:index]

	// add to the map
	attributes[name] = content
}

// newSectionData creates a new instance of SectionData and adds it to the 'sections' list.
// It uses a running number to generate the section ID in the format "sectionNNN".
func newSectionData(epubType, heading string) SectionData {
	currSectionNo++
	sectionId := fmt.Sprintf("section%03d", currSectionNo)
	section := SectionData{sectionId, epubType, heading}
	sections = append(sections, section)
	return section
}

// extractHeading extracts the plain text heading from the HTML tag <hx>...</x> where x is one of 1,2,3.
// On entry, currLine contains the string with the tag.
func extractHeading() string {
	var heading string
	if strings.HasPrefix(currLine, "<h1") || strings.HasPrefix(currLine, "<h2") || strings.HasPrefix(currLine, "<h3") {
		pos := strings.Index(currLine, ">") + 1
		heading = currLine[pos : len(currLine)-len("</h1>")]
	} else {
		fmt.Println("ERROR: Unexpected HTML line: ", currLine)
		fmt.Println("ERROR: HTML line with one of the tags <h1>, <h2> or <h3> expected")
		os.Exit(1)
	}
	if heading == "&#160;" {
		heading = ""
	}
	return heading
}
