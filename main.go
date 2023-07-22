// Copyright (C) 2022-2023, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Main source file

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/roslamir/ep3gen/internal/fileutil"
	"github.com/roslamir/ep3gen/internal/gen"
	"github.com/roslamir/ep3gen/internal/parm"
)

// Entry point
func main() {
	// Check arguments and load config parameters
	parm.CheckArgsAndParms(os.Args)

	// Loads the template files. Panics if any error occurs.
	gen.LoadTemplates()

	// Read in the whole input source file and store the lines in the string slice 'lines'.
	sourceDirSpec := filepath.Join(parm.SourceDir, parm.BookName)
	sourceFileSpec := filepath.Join(sourceDirSpec, "source.html")
	buffer := gen.NewInputBuffer(sourceFileSpec)

	// Remove the generated output directory and all children if it exists.
	targetDirSpec := filepath.Join(parm.TargetDir, parm.BookName)
	fileutil.DeleteDir(targetDirSpec)

	// Initialize the gen package
	gen.Init(sourceDirSpec, targetDirSpec)

	//-----------------------------------------------------------------------------------
	// Go through the source HTML lines and extract the metadata from the <head> section.
	//-----------------------------------------------------------------------------------

	// Skip over preliminary HTML lines until <head> is found
	for {
		buffer.NextLine()
		if buffer.CurrLine == "<head>" {
			break
		}
	}

	// Extract all the meta data defined and store them into the 'attributes' map.
	buffer.LoadAttributes()

	//-----------------------------------------------------------------------------------
	// Check for required attributes.
	//-----------------------------------------------------------------------------------

	var value string
	if value = buffer.GetAttribute("version"); value != "" {
		if value != "epub3" {
			panic("epubgen: attribute 'version' with value 'epub3' required")
		}
	} else {
		panic("epubgen: attribute 'version' required")
	}
	if value = buffer.GetAttribute("title"); value == "" {
		panic("epubgen: attribute 'title' required")
	}
	if value = buffer.GetAttribute("title-sort"); value == "" {
		panic("epubgen: attribute 'title-sort' required")
	}
	if value = buffer.GetAttribute("author"); value == "" {
		panic("epubgen: attribute 'author' required")
	}
	if value = buffer.GetAttribute("author-sort"); value == "" {
		panic("epubgen: attribute 'author-sort' required")
	}
	if value = buffer.GetAttribute("published"); value == "" {
		panic("epubgen: attribute 'published' required")
	}
	if value = buffer.GetAttribute("publisher"); value == "" {
		panic("epubgen: attribute 'publisher' required")
	}
	if value = buffer.GetAttribute("language"); value == "" {
		panic("epubgen: attribute 'language' required")
	}

	// Check and extract the mandatory attribute "cover-image" which specifies the cover image file.
	buffer.CheckCoverImage()

	// Check and extract the optional attribute "images" which lists all the image files embedded in the book other than the cover image.
	buffer.CheckImageFiles()

	// If updating an existing e-book, use the previous "created" attribute,
	// otherwise set the "created" attributes to the current timestamp.
	// In either case, set the "modified" attributes to the current timestamp.
	currTimeStamp := time.Now().UTC().Format(time.RFC3339)
	if value := buffer.GetAttribute("created"); value == "" {
		buffer.SetAttribute("created", currTimeStamp)
	}
	buffer.SetAttribute("modified", currTimeStamp)

	fmt.Printf("\nGenerating EPUB3 e-book \"%s\" from %s\n", buffer.GetAttribute("title"), parm.BookName)

	// Skip over the lines until the tag <body> is found
	for {
		buffer.NextLine()
		if buffer.CurrLine == "<body>" {
			break
		}
	}
	buffer.NextLine() // should point to the first directive

	//=============================
	// BOOK GENERATION STARTS HERE
	//=============================
	//------------------------------------------------------------------------
	// STEP 1: Generate the cover page section with data from the attributes.
	// Use the cover image file specified in the "cover-image" attribute.
	//------------------------------------------------------------------------
	buffer.GenCoverSection()

	//------------------------------------------------------------------------------------------------
	// Now, process the <body> section of the source HTML file. Lines containing HTML comments are
	// taken as directives in building the e-book. The last directive should be <!--end-->. Eveything
	// after it is ignored and it should be put just before the </body> tag.
	//------------------------------------------------------------------------------------------------

	//------------------------------------------------------------------------------------------------
	// STEP 2: Generate the title page section.
	//------------------------------------------------------------------------------------------------
	buffer.GenTitlePageSection()

	//------------------------------------------------------------------------------------------------
	// STEP 3: Generate the copyright section.
	// The next directive MUST be the "<!--copyright-->" section directive.
	//------------------------------------------------------------------------------------------------
	buffer.GenCopyrightSection(currTimeStamp[:10]) // Just use the date portion: 2006-01-02

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
		switch buffer.CurrLine {
		case "<!--bibliography-->":
			// Generate bibliography section, if requested.
			if bibliographyGiven {
				panic("epubgen: Directive <!--bibliography--> already specified")
			}
			bibliographyGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Bibliography"
			}
			section := buffer.NewSectionData("bibliography", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--acknowledgments-->":
			// Generate acknowledgments section, if requested.
			if acknowledgmentsGiven {
				panic("epubgen: Directive <!--acknowledgments--> already specified")
			}
			acknowledgmentsGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Acknowledgments"
			}
			section := buffer.NewSectionData("acknowledgments", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--dedication-->":
			// Generate dedication section, if requested.
			if dedicationGiven {
				panic("epubgen: Directive <!--dedication--> already specified")
			}
			dedicationGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Dedication"
			}
			section := buffer.NewSectionData("dedication", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--epigraph-->":
			// Generate epigraph section, if requested.
			if epigraphGiven {
				panic("epubgen: Directive <!--epigraph--> already specified")
			}
			epigraphGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Epigraph"
			}
			section := buffer.NewSectionData("epigraph", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--foreword-->":
			// Generate foreword section, if requested.
			if forewordGiven {
				panic("epubgen: Directive <!--foreword--> already specified")
			}
			forewordGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Foreword"
			}
			section := buffer.NewSectionData("foreword", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--introduction-->":
			// Generate introduction section, if requested.
			if introductionGiven {
				panic("epubgen: Directive <!--introduction--> already specified")
			}
			introductionGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Introduction"
			}
			section := buffer.NewSectionData("introduction", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--preface-->":
			// Generate preface section, if requested.
			if prefaceGiven {
				panic("epubgen: Directive <!--preface--> already specified")
			}
			prefaceGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Preface"
			}
			section := buffer.NewSectionData("preface", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--prologue-->":
			// Generate prologue section, if requested.
			if prologueGiven {
				panic("epubgen: Directive <!--prologue--> already specified")
			}
			prologueGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Prologue"
			}
			section := buffer.NewSectionData("prologue", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

		case "<!--preamble-->":
			// Generate generic preamble section, may occur multiple times.
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Preamble"
			}
			section := buffer.NewSectionData("preamble", heading)
			buffer.AddSection(section)
			buffer.GenFrontMatterSection(section)

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
		switch buffer.CurrLine {
		case "<!--part-->":
			// Generate part section, may occur zero or more times
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			section := buffer.NewSectionData("part", heading)
			buffer.AddSection(section)
			buffer.GenBodyMatterSection(section)
			if firstBodymatter {
				firstBodymatter = false
				buffer.AddGuide(section) // add to guides slice
			}

		case "<!--chapter-->":
			// Generate chapter section, may occur one or more times
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			section := buffer.NewSectionData("chapter", heading)
			buffer.AddSection(section)
			buffer.GenBodyMatterSection(section)
			if firstBodymatter {
				firstBodymatter = false
				buffer.AddGuide(section) // add to guides slice
			}

		default:
			break loop2
		}
	}

	// If the flag 'firstBodymatter' is still true, it means neither part nor chapter was given, and
	// we treat this as an error condition.
	if firstBodymatter {
		panic("epubgen: at least one <!--chapter--> directive must be specified")
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
		switch buffer.CurrLine {
		case "<!--afterword-->":
			// Generate afterword section, if specified.
			if afterwordGiven {
				panic("epubgen: Directive <!--afterword--> already specified")
			}
			afterwordGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Afterword"
			}
			section := buffer.NewSectionData("afterword", heading)
			buffer.AddSection(section)
			buffer.GenBackMatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				buffer.AddGuide(section)
			}

		case "<!--epilogue-->":
			// Generate epilogue section, if specified.
			if epilogueGiven {
				panic("epubgen: Directive <!--epilogue--> already specified")
			}
			epilogueGiven = true
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Epilogue"
			}
			section := buffer.NewSectionData("epilogue", heading)
			buffer.AddSection(section)
			buffer.GenBackMatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				buffer.AddGuide(section)
			}

		case "<!--appendix-->":
			// Generate appendix section if specified, may occur multiple times.
			buffer.NextLine()
			heading := extractHeading(buffer.CurrLine)
			if heading == "" {
				heading = "Appendix"
			}
			section := buffer.NewSectionData("appendix", heading)
			buffer.AddSection(section)
			buffer.GenBackMatterSection(section)
			if firstBackmatter {
				firstBackmatter = false
				buffer.AddGuide(section)
			}

		case "<!--end-->":
			break loop3

		default:
			panic("epubgen: Unknown directive: " + buffer.CurrLine)
		}
	}

	//------------------------------------------------------------------------------------------------
	// STEP 7: Generate the control files (nav.xhtml, toc.ncx and package.opf)
	//------------------------------------------------------------------------------------------------

	// Generate NAV (TOC) file (required for EPUB3)
	buffer.GenNAVFile()

	// Generate NCX file (for EPUB2 compatibility)
	buffer.GenNCXFile()

	// Generate the package (OPF) file
	buffer.GenOPFFile()

	//------------------------------------------------------------------------------------------------
	// STEP 8: Copy the static (resource and image) files unchanged.
	//------------------------------------------------------------------------------------------------

	// Copy the control files, the stylesheet and the image files
	buffer.CopyStaticFiles()

	fmt.Printf("\n%d lines processed\n", buffer.NumLines())
}

// extractMetaData extracts the metadata 'name' and 'content' from the current line.
// Returns the name and content of the metadata.
// func extractMetaData(line string) (string, string) {
// 	index := strings.Index(line, "name=")
// 	if index == -1 {
// 		return "", ""
// 	}
// 	name := line[index+len("name=")+1:] // skip past 'name="'
// 	index = strings.Index(name, "\"")
// 	if index == -1 {
// 		panic("epubgen: Invalid 'meta' HTML line: ", line)
// 	}
// 	name = name[:index]

// 	index = strings.Index(line, "content=")
// 	if index == -1 {
// 		panic("epubgen: Invalid 'meta' HTML line: ", line)
// 	}
// 	content := line[index+len("content=")+1:] // skip past 'content="'
// 	index = strings.Index(content, "\"")
// 	if index == -1 {
// 		panic("epubgen: Invalid 'meta' HTML line: ", line)
// 	}
// 	content = content[:index]

// 	return name, content
// }

// extractHeading extracts the plain text heading from the HTML tag <hx>...</x> where x is one of 1,2,3.
// On entry, line contains the string with the tag.
func extractHeading(line string) string {
	var heading string
	if strings.HasPrefix(line, "<h1") || strings.HasPrefix(line, "<h2") || strings.HasPrefix(line, "<h3") {
		pos := strings.Index(line, ">") + 1
		heading = line[pos : len(line)-5] // 5 is the length of </hN>
	} else {
		panic("epubgen: HTML line with one of the tags <h1>, <h2> or <h3> expected")
	}
	if heading == "&#160;" {
		heading = ""
	}
	return heading
}
