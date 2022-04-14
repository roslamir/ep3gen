// Copyright (C) 2022, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Collection of file utility functions.

package fileutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadLines reads in the input source file line by line and store in an array of lines.
// Input: string representing the file spec.
// Output: []string - array of strings containing the lines from the file (each line stripped off '\n')
func ReadLines(fileSpec string) []string {
	var file *os.File
	var err error

	if file, err = os.Open(fileSpec); err != nil {
		fmt.Println("Open failed:", err)
		os.Exit(255)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	return lines
}

// CreateFile creates output file given the file spec.
// Return the pointer to an os.File instance.
func CreateFile(filespec string) *os.File {
	var file *os.File
	var err error

	if file, err = os.Create(filespec); err != nil {
		fmt.Println("Create failed:", err)
		os.Exit(255)
	}

	return file
}

// DeleteDir removes the specified directory and all children if it exists.
func DeleteDir(dirspec string) {
	if err := os.RemoveAll(dirspec); err != nil {
		fmt.Println("RemoveAll failed:", err)
		os.Exit(255)
	}
}

// CreateDirs creates the specified directory and all its parents if necessary.
func CreateDirs(dirspec string) {
	if err := os.MkdirAll(dirspec, os.ModePerm); err != nil {
		fmt.Println("MkdirAll failed:", err)
		os.Exit(255)
	}
}

// FileCopy copies the source file to the target file, overwriting if needed.
func FileCopy(sourcefilespec, targetfilespec string) {
	var infile *os.File
	var err error
	if infile, err = os.Open(sourcefilespec); err != nil {
		fmt.Println("Open failed:", err)
		os.Exit(255)
	}
	defer infile.Close()

	// to, err := os.OpenFile(targetfilespec, os.O_RDWR|os.O_CREATE, 0666)
	var outfile *os.File
	if outfile, err = os.Create(targetfilespec); err != nil {
		fmt.Println("Create failed:", err)
		os.Exit(255)
	}
	defer outfile.Close()

	if _, err = io.Copy(outfile, infile); err != nil {
		fmt.Println("Copy failed:", err)
		os.Exit(255)
	}
}
