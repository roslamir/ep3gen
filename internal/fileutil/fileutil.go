// Copyright (C) 2022, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Collection of file utility functions.

package fileutil

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DeleteDir removes the specified directory and all children if it exists.
func DeleteDir(dirspec string) {
	if err := os.RemoveAll(dirspec); err != nil {
		panic(err)
	}
}

// OpenFile opens input file for reading given the file spec.
func OpenFile(fileSpec string) *os.File {
	var err error
	var file *os.File
	if file, err = os.Open(fileSpec); err != nil {
		panic(err)
	}
	return file
}

// CreateFile creates output file given the file spec.
// Also creates any parent directory along the path if necessary.
func CreateFile(filespec string) *os.File {
	var err error
	if err = os.MkdirAll(filepath.Dir(filespec), 0770); err != nil {
		panic(err)
	}
	var file *os.File
	if file, err = os.Create(filespec); err != nil {
		panic(err)
	}
	return file
}

// ReadLines reads in the input source file line by line and store in an array of lines.
// Input: string representing the file spec.
// Output: []string - array of strings containing the lines from the file (each line stripped off '\n')
func ReadLines(sourcefilespec string) []string {
	// Open source file for reading
	infile := OpenFile(sourcefilespec)
	defer infile.Close()

	scanner := bufio.NewScanner(infile)
	scanner.Split(bufio.ScanLines)
	lines := make([]string, 0, 1024)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	return lines
}

// CopyFile copies the source file to the target file, overwriting if needed.
func CopyFile(sourcefilespec, targetfilespec string) {
	// Open source file for reading
	infile := OpenFile(sourcefilespec)
	defer infile.Close()

	// Create output file for writing
	outfile := CreateFile(targetfilespec)
	defer outfile.Close()

	// Copy the source file to the target file
	if _, err := io.Copy(outfile, infile); err != nil {
		panic(err)
	}
}
