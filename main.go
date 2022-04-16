// Copyright (C) 2022, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Main source file

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/google/uuid"
	"github.com/roslamir/ep3gen/internal/fileutil"
)

const (
	coverTemplate            = "cover.gohtml"
	defaultTitlepageTemplate = "default-titlepage.gohtml"
	imageTitlepageTemplate   = "image-titlepage.gohtml"
	frontmatterTemplate      = "frontmatter.gohtml"
	bodymatterTemplate       = "bodymatter.gohtml"
	backmatterTemplate       = "backmatter.gohtml"
	navTemplate              = "nav.gohtml"
	ncxTemplate              = "ncx.goxml"
	opfTemplate              = "opf.goxml"
)

// SectionData holds the attributes for a section.
// Each generated HTML is considered a section and each section metadata is kept here.
type SectionData struct {
	ID       string // section id is used as the name of the section file and also used as the id in the package manifest
	EpubType string // used as the value for "epub-type" attribute for the HTML <section> tag
	Heading  string // used as the section heading to be displayed in the table of contents (TOC)
}

// ImageData holds the name and extension for an image file.
type ImageData struct {
	Name string // image file name without extension
	Ext  string // extension also serves as the media type (png/jpeg)
}

var (
	sourceDir    = "./data/source"
	targetDir    = "./data/generated"
	resourceDir  = "./data/etc"
	templatesDir = "./data/templates"

	tmpl           *template.Template
	bookName       string            // the directory under which all the source files are located
	packageDirSpec string            // the full path for the OEBPS directory
	textDirSpec    string            // the full path for the OEBPS/Text directory
	lines          []string          // holds the list of all lines from the source HTML file
	currLine       string            // holds the string representing the current line
	lineIndex      int               // index into the 'lines' slice', points to the current line
	attributes     map[string]string // contains all the metadata attibutes
	coverImage     ImageData         // holds the file name and extension for the cover image
	images         []ImageData       // holds the list of all image files (other than the cover image) used in the book
	sections       []SectionData     = make([]SectionData, 0)
	guides         []SectionData     = make([]SectionData, 0)               // used in the Guides section of the manifest
	bookUUID       string            = strings.ToUpper(uuid.New().String()) // Always create a new UUID for this e-book
	currSectionNo  int                                                      // Holds the current section counter
)

func init() {
	// Read in the configuration values
	if cfgfile, err := os.ReadFile("./config.yaml"); err == nil {
		cfgMap := make(map[string]string)
		err = yaml.Unmarshal(cfgfile, &cfgMap)
		if err != nil {
			fmt.Println("WARNING: Error unmarshalling config.yaml file", err)
			os.Exit(1)
		}
		if value, exists := cfgMap["source_dir"]; exists {
			sourceDir = value
		}
		if value, exists := cfgMap["target_dir"]; exists {
			targetDir = value
		}
		if value, exists := cfgMap["resource_dir"]; exists {
			resourceDir = value
		}
		if value, exists := cfgMap["templates_dir"]; exists {
			templatesDir = value
		}
	} else {
		fmt.Println("WARNING: Cannot read config.yaml file -- using defaults")
	}

	// Checks if any of the above is overriden by an environment variable
	if value := os.Getenv("EP3GEN_SOURCE_DIR"); value != "" {
		sourceDir = value
	}
	if value := os.Getenv("EP3GEN_TARGET_DIR"); value != "" {
		targetDir = value
	}
	if value := os.Getenv("EP3GEN_RESOURCE_DIR"); value != "" {
		resourceDir = value
	}
	if value := os.Getenv("EP3GEN_TEMPLATES_DIR"); value != "" {
		templatesDir = value
	}

	// Load in the template files
	tmpl = template.Must(template.ParseFiles(
		filepath.Join(templatesDir, coverTemplate),
		filepath.Join(templatesDir, defaultTitlepageTemplate),
		filepath.Join(templatesDir, imageTitlepageTemplate),
		filepath.Join(templatesDir, frontmatterTemplate),
		filepath.Join(templatesDir, bodymatterTemplate),
		filepath.Join(templatesDir, backmatterTemplate),
		filepath.Join(templatesDir, navTemplate),
		filepath.Join(templatesDir, ncxTemplate),
		filepath.Join(templatesDir, opfTemplate),
	))
}

// Entry point
func main() {
	// Check that the bookdir argument is given
	checkArgs()

	// Read in the whole input source file and store the lines in the string slice 'lines'.
	sourceDirSpec := filepath.Join(sourceDir, bookName)
	sourceFileSpec := filepath.Join(sourceDirSpec, "source.html")
	lines = fileutil.ReadLines(sourceFileSpec)

	// Remove the generated output directory and all children if it exists.
	targetDirSpec := filepath.Join(targetDir, bookName)
	fileutil.DeleteDir(targetDirSpec)

	// Create the EPUB directory tree
	metainfDirSpec := filepath.Join(targetDirSpec, "META-INF")
	fileutil.CreateDirs(metainfDirSpec)
	packageDirSpec = filepath.Join(targetDirSpec, "OEBPS")
	imagesDirSpec := filepath.Join(packageDirSpec, "Images")
	fileutil.CreateDirs(imagesDirSpec)
	stylesDirSpec := filepath.Join(packageDirSpec, "Styles")
	fileutil.CreateDirs(stylesDirSpec)
	textDirSpec = filepath.Join(packageDirSpec, "Text")
	fileutil.CreateDirs(textDirSpec)

	//-----------------------------------------------------------------------------------
	// Go through the source HTML lines and extract the metadata from the <head> section.
	//-----------------------------------------------------------------------------------

	// Skip over preliminary HTML lines until <head> is found
	for {
		currLine = nextLine()
		if currLine == "<head>" {
			break
		}
	}

	// Extract all the meta data defined and store them into the 'attributes' map.
	attributes = make(map[string]string)
	for {
		currLine = nextLine()
		if currLine == "</head>" {
			break
		}
		if strings.HasPrefix(currLine, "<meta") {
			extractMetaData()
		}
	}

	//-----------------------------------------------------------------------------------
	// Check for required attributes.
	//-----------------------------------------------------------------------------------

	var exists bool
	var attr string
	if attr, exists = attributes["version"]; exists {
		if attr != "epub3" {
			fmt.Println("ERROR: Attribute 'version' with value 'epub3' required")
			os.Exit(1)
		}
	} else {
		fmt.Println("ERROR: Attribute 'version' required")
		os.Exit(1)
	}
	if _, exists = attributes["title"]; !exists {
		fmt.Println("ERROR: Attribute 'title' required")
		os.Exit(1)
	}
	if _, exists = attributes["title-sort"]; !exists {
		fmt.Println("ERROR: Attribute 'title-sort' required")
		os.Exit(1)
	}
	if _, exists = attributes["author"]; !exists {
		fmt.Println("ERROR: Attribute 'author' required")
		os.Exit(1)
	}
	if _, exists = attributes["author-sort"]; !exists {
		fmt.Println("ERROR: Attribute 'author-sort' required")
		os.Exit(1)
	}
	if _, exists = attributes["published"]; !exists {
		fmt.Println("ERROR: Attribute 'published' required")
		os.Exit(1)
	}
	if _, exists = attributes["publisher"]; !exists {
		fmt.Println("ERROR: Attribute 'publisher' required")
		os.Exit(1)
	}
	if _, exists = attributes["language"]; !exists {
		fmt.Println("ERROR: Attribute 'language' required")
		os.Exit(1)
	}

	// Attribute "cover-image" must be present.
	// The value must be the name of the cover image file with extension of either ".jpeg" or ".png".
	// To make life easier, assume all JPEG files have extension ".jpeg" instead of ".jpg".
	if coverImageFile, exists := attributes["cover-image"]; exists {
		parts := strings.Split(coverImageFile, ".")
		coverImage = ImageData{
			Name: parts[0],
			Ext:  parts[1],
		}
		if coverImage.Ext != "png" && coverImage.Ext != "jpeg" {
			fmt.Println("ERROR: Only image files with extension 'png' or 'jpeg' are accepted")
			os.Exit(1)
		}
	} else {
		fmt.Println("ERROR: Attribute 'cover-image' required")
		os.Exit(1)
	}

	// All image files used in the book must be listed in the attribute "images" separated by commas without spaces.
	if value, exists := attributes["images"]; exists {
		imageFiles := strings.Split(value, ",")
		for _, imageFile := range imageFiles {
			parts := strings.Split(imageFile, ".")
			image := ImageData{
				Name: parts[0],
				Ext:  parts[1],
			}
			if image.Ext != "png" && image.Ext != "jpeg" {
				fmt.Println("ERROR: Only image files with extension 'png' or 'jpeg' are accepted")
				os.Exit(1)
			}
			images = append(images, image)
		}
	}

	// If updating an existing e-book, use the previous "created" attribute,
	// otherwise set the "created" attributes to the current timestamp.
	// In either case, set the "modified" attributes to the current timestamp.
	currTimeStamp := time.Now().UTC().Format(time.RFC3339)
	if _, exists := attributes["created"]; !exists {
		attributes["created"] = currTimeStamp
	}
	attributes["modified"] = currTimeStamp

	fmt.Printf("\nGenerating EPUB3 e-book \"%s\" from %s\n", attributes["title"], bookName)

	// Skip over the lines until <body> is found
	for {
		currLine = nextLine()
		if currLine == "<body>" {
			break
		}
	}
	currLine = nextLine() // should point to the first directive

	//--------------------------------------------------------------------
	// STEP 1: Generate the cover page section.
	// Use the cover image file specified in the "cover-image" attribute.
	//--------------------------------------------------------------------
	section := SectionData{"cover", "cover", "Cover Page"}
	sections = append(sections, section)
	guides = append(guides, section)
	genCoverSection(section)

	//------------------------------------------------------------------------------------------------
	// Now, process the <body> section of the source HTML file. Lines containing HTML comments are
	// taken as directives in building the e-book. The last directive should be <!--end-->. Eveything
	// after it is ignored and it should be put just before the </body> tag.
	//------------------------------------------------------------------------------------------------

	//------------------------------------------------------------------------------------------------
	// STEP 2: Generate the title page section.
	//------------------------------------------------------------------------------------------------

	// If the attribute "titlepage" is not given or has the value of "default", we generate the
	// default title page section.
	// If the attribute "titlepage" has the value of "custom", the first directive encountered must be
	// "<!--titlepage-->" and it must be followed by one or more formatted HTML lines making up the
	// title page section.
	// Otherwise, we assume that the value is the name of an image file with either "png" or "jpeg"
	// extension.
	var titlePage string
	if titlePage, exists = attributes["titlepage"]; !exists {
		titlePage = "default"
	}

	section = SectionData{"titlepage", "titlepage", "Title Page"}
	sections = append(sections, section)
	guides = append(guides, section)

	switch titlePage {
	case "default":
		genDefaultTitlePageSection(section)

	case "custom":
		currLine = nextLine()
		if currLine == "<!--titlepage-->" {
			currLine = nextLine()
			genFrontmatterSection(section)
		} else {
			fmt.Println("ERROR: <!--titlepage--> directive expected")
			os.Exit(1)
		}

	default: // assumes the value is an image file for the title page
		parts := strings.Split(titlePage, ".")
		image := ImageData{
			Name: parts[0],
			Ext:  parts[1],
		}
		if image.Ext != "png" && image.Ext != "jpeg" {
			fmt.Println("ERROR: Only image files with extension 'png' or 'jpeg' are accepted")
			os.Exit(1)
		}
		genImageTitlePageSection(section, image)
	}

	//------------------------------------------------------------------------------------------------
	// STEP 3: Generate the copyright section.
	//------------------------------------------------------------------------------------------------

	// The next directive MUST be the "<!--copyright-->" section directive.
	if currLine == "<!--copyright-->" {
		currLine = nextLine()
		section = SectionData{"copyright", "copyright-page", "Copyright"}
		sections = append(sections, section)
		genFrontmatterSection(section)
	} else {
		fmt.Println("ERROR: <!--copyright--> directive expected")
		os.Exit(1)
	}

	//------------------------------------------------------------------------------------------------
	// STEP 4: Generate the optional frontmatter sections.
	// The optional fontmatter directives are:
	// 1. <!--bibliography-->
	// 2. <!--acknowledgments-->
	// 3. <!--dedication-->
	// 4. <!--epigraph-->
	// 5. <!--foreword-->
	// 6. <!--introduction-->
	// 7. <!--prologue-->
	// 8. <!--preamble-->
	// The first seven may only occur once but 'preamble' may occur multiple times as a generic
	// frontmatter section not covered by the first seven.
	// The first line after the directive must be the section heading formatted as one of the HTML
	// tags: <h1>, <h2> or <h3>.
	// If no heading is applicable, use '<h1>&#160;</h1>' for the heading line.
	// It must be followed by one or more formatted HTML lines making up the frontmatter section.
	// Each directive must be followed by one of <h1>, <h2> or <h3> tags with the section heading.
	// If no heading is needed, Use <h1>&#160;</h1> and the default heading will be used in the TOC.
	//------------------------------------------------------------------------------------------------

	var (
		bibliographyGiven    bool
		acknowledgmentsGiven bool
		dedicationGiven      bool
		epigraphGiven        bool
		forewordGiven        bool
		introductionGiven    bool
		prefaceGiven         bool
		prologueGiven        bool
	)

loop1:
	for {
		switch currLine {
		case "<!--bibliography-->":
			// Generate bibliography section, if requested.
			if bibliographyGiven {
				fmt.Println("ERROR: Directive <!--bibliography--> already specified")
				os.Exit(1)
			}
			bibliographyGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Bibliography"
			}
			section = newSectionData("bibliography", heading)
			genFrontmatterSection(section)

		case "<!--acknowledgments-->":
			// Generate acknowledgments section, if requested.
			if acknowledgmentsGiven {
				fmt.Println("ERROR: Directive <!--acknowledgments--> already specified")
				os.Exit(1)
			}
			acknowledgmentsGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Acknowledgments"
			}
			section = newSectionData("acknowledgments", heading)
			genFrontmatterSection(section)

		case "<!--dedication-->":
			// Generate dedication section, if requested.
			if dedicationGiven {
				fmt.Println("ERROR: Directive <!--dedication--> already specified")
				os.Exit(1)
			}
			dedicationGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Dedication"
			}
			section = newSectionData("dedication", heading)
			genFrontmatterSection(section)

		case "<!--epigraph-->":
			// Generate epigraph section, if requested.
			if epigraphGiven {
				fmt.Println("ERROR: Directive <!--epigraph--> already specified")
				os.Exit(1)
			}
			epigraphGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Epigraph"
			}
			section = newSectionData("epigraph", heading)
			genFrontmatterSection(section)

		case "<!--foreword-->":
			// Generate foreword section, if requested.
			if forewordGiven {
				fmt.Println("ERROR: Directive <!--foreword--> already specified")
				os.Exit(1)
			}
			forewordGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Foreword"
			}
			section = newSectionData("foreword", heading)
			genFrontmatterSection(section)

		case "<!--introduction-->":
			// Generate introduction section, if requested.
			if introductionGiven {
				fmt.Println("ERROR: Directive <!--introduction--> already specified")
				os.Exit(1)
			}
			introductionGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Introduction"
			}
			section = newSectionData("introduction", heading)
			genFrontmatterSection(section)

		case "<!--preface-->":
			// Generate preface section, if requested.
			if prefaceGiven {
				fmt.Println("ERROR: Directive <!--preface--> already specified")
				os.Exit(1)
			}
			prefaceGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Preface"
			}
			section = newSectionData("preface", heading)
			genFrontmatterSection(section)

		case "<!--prologue-->":
			// Generate prologue section, if requested.
			if prologueGiven {
				fmt.Println("ERROR: Directive <!--prologue--> already specified")
				os.Exit(1)
			}
			prologueGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Prologue"
			}
			section = newSectionData("prologue", "Prologue")
			genFrontmatterSection(section)

		case "<!--preamble-->":
			// Generate generic preamble section, may occur multiple times.
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Preamble"
			}
			section = newSectionData("preamble", "Preamble")
			genFrontmatterSection(section)

		default:
			break loop1
		}
	}

	//------------------------------------------------------------------------------------------------
	// STEP 5: Generate the part and chapter (bodymatter) sections.
	// An e-book may consist of zero or more parts and one or more chapters.
	// We also check if the part or chapter is the first since we want to add that section to the
	// Guides page for the book.
	//------------------------------------------------------------------------------------------------

	firstBodymatter := true

loop2:
	for {
		switch currLine {
		case "<!--part-->":
			// Generate part section, may occur zero or more times
			currLine = nextLine()
			heading := extractHeading()
			section = newSectionData("part", heading)
			genBodyMatterSection(section)
			if firstBodymatter {
				firstBodymatter = false
				guides = append(guides, section) // add to guides slice
			}

		case "<!--chapter-->":
			// Generate chapter section, may occur one or more times
			currLine = nextLine()
			heading := extractHeading()
			section = newSectionData("chapter", heading)
			genBodyMatterSection(section)
			if firstBodymatter {
				firstBodymatter = false
				guides = append(guides, section) // add to guides slice
			}

		default:
			break loop2
		}
	}

	// If the flag 'firstBodymatter' is still true, it means neither part nor chapter was given, and
	// we treat this as an error condition.
	if firstBodymatter {
		fmt.Println("ERROR: At least one <!--chapter--> directive must be specified")
		os.Exit(1)
	}

	//------------------------------------------------------------------------------------------------
	// STEP 6: Generate the optional backmatter sections.
	// The optional fontmatter directives are:
	// 1. <!--afterword-->
	// 2. <!--epilogue-->
	// 3. <!--appendix-->
	// The first two may only occur once but 'appendix' may occur multiple times as a generic
	// backmatter section not covered by the first two.
	// The first line after the directive must be the section heading formatted as one of the HTML
	// tags: <h1>, <h2> or <h3>.
	// If no heading is applicable, use '<h1>&#160;</h1>' for the heading line.
	// It must be followed by one or more formatted HTML lines making up the backmatter section.
	//------------------------------------------------------------------------------------------------

	var (
		afterwordGiven  bool
		epilogueGiven   bool
		firstBackmatter bool = true
	)

loop3:
	for {
		switch currLine {
		case "<!--afterword-->":
			// Generate afterword section, if specified.
			if afterwordGiven {
				fmt.Println("ERROR: Directive <!--afterword--> already specified")
				os.Exit(1)
			}
			afterwordGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Afterword"
			}
			section = newSectionData("afterword", heading)
			genBackmatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				guides = append(guides, section)
			}

		case "<!--epilogue-->":
			// Generate epilogue section, if specified.
			if epilogueGiven {
				fmt.Println("ERROR: Directive <!--epilogue--> already specified")
				os.Exit(1)
			}
			epilogueGiven = true
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Epilogue"
			}
			section = newSectionData("epilogue", heading)
			genBackmatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				guides = append(guides, section)
			}

		case "<!--appendix-->":
			// Generate appendix section if specified, may occur multiple times.
			currLine = nextLine()
			heading := extractHeading()
			if heading == "" {
				heading = "Appendix"
			}
			section = newSectionData("appendix", heading)
			genBackmatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				guides = append(guides, section)
			}

		case "<!--end-->":
			break loop3

		default:
			fmt.Println("ERROR: Unknown directive:", currLine)
			os.Exit(1)
		}
	}

	//------------------------------------------------------------------------------------------------
	// STEP 7: Generate the control files (nav.xhtml, toc.ncx and package.opf)
	//------------------------------------------------------------------------------------------------

	// Generate NAV (TOC) file (required for EPUB3)
	genNAVFile()

	// Generate NCX file (for EPUB2 compatibility)
	genNCXFile()

	// Generate the package (OPF) file
	genOPFFile()

	//------------------------------------------------------------------------------------------------
	// STEP 8: Copy the static (resource and image) files unchanged.
	//------------------------------------------------------------------------------------------------

	// <bookdir>/mimetype
	sourceFileSpec = filepath.Join(resourceDir, "mimetype")
	targetFileSpec := filepath.Join(targetDirSpec, "mimetype")
	fileutil.FileCopy(sourceFileSpec, targetFileSpec)

	// <bookdir>/META-INF/container.xml
	sourceFileSpec = filepath.Join(resourceDir, "container.xml")
	targetFileSpec = filepath.Join(metainfDirSpec, "container.xml")
	fileutil.FileCopy(sourceFileSpec, targetFileSpec)

	// <bookdir>/OEBPS/Styles/stylesheet.css
	sourceFileSpec = filepath.Join(resourceDir, "stylesheet.css")
	targetFileSpec = filepath.Join(stylesDirSpec, "stylesheet.css")
	fileutil.FileCopy(sourceFileSpec, targetFileSpec)

	// <bookdir>/OEBPS/Images/*
	coverImageFile := coverImage.Name + "." + coverImage.Ext
	sourceFileSpec = filepath.Join(sourceDirSpec, coverImageFile)
	targetFileSpec = filepath.Join(imagesDirSpec, coverImageFile)
	fileutil.FileCopy(sourceFileSpec, targetFileSpec)
	if value, exists := attributes["images"]; exists {
		imageFiles := strings.Split(value, ",")
		for _, imageFile := range imageFiles {
			sourceFileSpec = filepath.Join(sourceDirSpec, imageFile)
			targetFileSpec = filepath.Join(imagesDirSpec, imageFile)
			fileutil.FileCopy(sourceFileSpec, targetFileSpec)
		}
	}

	fmt.Printf("\n%d lines processed\n", len(lines))
}
