// Copyright (C) 2022-2023, Roslan Amir. All rights reserved.
// Created on: 21-Jul-2023
//
// Check arguments and load configuration parameters

package parm

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	usage = `usage: epubgen [-c path_to_config_file] BookName

Generates EPUB3 e-book from the source artifacts under the directory ./source/<BookName>.`
)

var (
	BookUUID     string = strings.ToUpper(uuid.New().String()) // Always create a new UUID for this e-book
	BookName     string
	SourceDir    string
	TargetDir    string
	ResourceDir  string
	TemplatesDir string
)

// checkArgs checks the input arguments and acts accordingly.
func CheckArgsAndParms(args []string) {
	var configFile string
	if len(args) == 1 {
		// Show usage information if no arguments are given
		fmt.Println(usage)
		os.Exit(1)
	} else if len(args) == 2 {
		// Assume only the 'BookName' is given
		BookName = args[1]
	} else if len(args) == 4 {
		if args[1] == "-c" {
			configFile = args[2]
			BookName = args[3]
		} else {
			fmt.Println(usage)
			os.Exit(1)
		}
	} else {
		// Show usage information if extraneous arguments are given
		fmt.Println(usage)
		os.Exit(1)
	}

	// Use default config file
	if configFile == "" {
		configFile = "./config.yaml"
	}

	// Read in the configuration values
	if cfgfile, err := os.ReadFile(configFile); err == nil {
		cfgMap := make(map[string]string)
		err = yaml.Unmarshal(cfgfile, &cfgMap)
		if err != nil {
			msg := fmt.Sprintf("epubgen: error unmarshalling config file %s: %s", configFile, err.Error())
			panic(msg)
		}
		if value, exists := cfgMap["source_dir"]; exists {
			SourceDir = value
		} else {
			msg := fmt.Sprintf("epubgen: config parameter '%s' required", "source_dir")
			panic(msg)
		}
		if value, exists := cfgMap["target_dir"]; exists {
			TargetDir = value
		} else {
			msg := fmt.Sprintf("epubgen: config parameter '%s' required", "target_dir")
			panic(msg)
		}
		if value, exists := cfgMap["resource_dir"]; exists {
			ResourceDir = value
		} else {
			msg := fmt.Sprintf("epubgen: config parameter '%s' required", "resource_dir")
			panic(msg)
		}
		if value, exists := cfgMap["templates_dir"]; exists {
			TemplatesDir = value
		} else {
			msg := fmt.Sprintf("epubgen: config parameter '%s' required", "templates_dir")
			panic(msg)
		}
	} else {
		msg := fmt.Sprintf("epubgen: cannot read config file %s: %s", configFile, err.Error())
		panic(msg)
	}
}
