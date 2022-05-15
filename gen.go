// Copyright (C) 2022, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Functions to generate the various sections of the e-book

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roslamir/ep3gen/internal/fileutil"
)

// genCoverSection generates the cover page section.
func genCoverSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		Title      string
		CoverImage ImageData
	}{
		Title:      attributes["title"],
		CoverImage: coverImage,
	}

	if err := tmpl.ExecuteTemplate(outfile, coverTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genDefaultTitlePageSection generates the default title page section.
func genDefaultTitlePageSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Struct to pass to the template
	_, hasSubtitle := attributes["subtitle"]
	_, hasSeries := attributes["series"]
	_, hasAuthor2 := attributes["author2"]
	_, hasAuthor3 := attributes["author3"]
	data := struct {
		Title       string
		HasSubtitle bool
		Subtitle    string
		HasSeries   bool
		Series      string
		SeriesIndex string
		Author      string
		HasAuthor2  bool
		Author2     string
		HasAuthor3  bool
		Author3     string
		Publisher   string
		Published   string
	}{
		Title:       attributes["title"],
		HasSubtitle: hasSubtitle,
		Subtitle:    attributes["subtitle"],
		HasSeries:   hasSeries,
		Series:      attributes["series"],
		SeriesIndex: attributes["series-index"],
		Author:      attributes["author"],
		HasAuthor2:  hasAuthor2,
		Author2:     attributes["author2"],
		HasAuthor3:  hasAuthor3,
		Author3:     attributes["author3"],
		Publisher:   attributes["publisher"],
		Published:   attributes["published"],
	}

	if err := tmpl.ExecuteTemplate(outfile, defaultTitlepageTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genImageTitlePageSection generates the title page section comprising a single image.
func genImageTitlePageSection(section SectionData, image ImageData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Image    ImageData
		Heading  string
	}{
		Title:    attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Image:    image,
		Heading:  section.Heading,
	}

	if err := tmpl.ExecuteTemplate(outfile, imageTitlepageTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genFrontmatterSection generates one of the various frontmatter sections file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func genFrontmatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0)
	for {
		sectionLines = append(sectionLines, currLine)
		currLine = nextLine()
		if strings.HasPrefix(currLine, "<!--") {
			break
		}
	}

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Lines    []string
	}{
		Title:    attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, frontmatterTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genBodyMatterSection generates the bodymatter (part or chapter) section file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func genBodyMatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0)
	for {
		sectionLines = append(sectionLines, currLine)
		currLine = nextLine()
		if strings.HasPrefix(currLine, "<!--") {
			break
		}
	}

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Lines    []string
	}{
		Title:    attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, bodymatterTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genBackmatterSection generates the copyright section file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func genBackmatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0)
	for {
		sectionLines = append(sectionLines, currLine)
		currLine = nextLine()
		if strings.HasPrefix(currLine, "<!--") {
			break
		}
	}

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Lines    []string
	}{
		Title:    attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, backmatterTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// PartSectionData holds the list of part sections with their chapter sections.
type PartSectionData struct {
	Part     SectionData
	Chapters []SectionData
}

// genNAVFile generates the NAV (TOC) file (required for EPUB3).
func genNAVFile() {
	fileName := "nav.xhtml"
	fmt.Printf("Generating file %s (TOC) ... ", fileName)

	outfilespec := filepath.Join(textDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	var index int
	var section SectionData

	// Get the slice of 'sections' that forms the frontmatter
	for index, section = range sections {
		if section.EpubType == "part" || section.EpubType == "chapter" {
			break
		}
	}
	frontSections := sections[:index]

	// Flag indicating whether this book contains parts and chapters or just chapters
	hasParts := section.EpubType == "part"

	var partSections []PartSectionData
	var chapterSections []SectionData
	var startIndex int
	if hasParts {
		// Get the slice of 'sections' that forms the parts and chapters
		firstTime := true
		var currPart SectionData
		for {
			section = sections[index]
			if section.EpubType == "part" {
				if firstTime {
					firstTime = false
					partSections = make([]PartSectionData, 0)
					currPart = section
					startIndex = index + 1
				} else {
					partSection := PartSectionData{
						Part:     currPart,
						Chapters: sections[startIndex:index],
					}
					partSections = append(partSections, partSection)
					currPart = section
					startIndex = index + 1
				}
			} else if section.EpubType != "chapter" {
				partSection := PartSectionData{
					Part:     currPart,
					Chapters: sections[startIndex:index],
				}
				partSections = append(partSections, partSection)
				break
			}
			index++
		}
	} else {
		// Get the slice of 'sections' that forms the chapters (no parts)
		startIndex := index
		for ; sections[index].EpubType != "chapter"; index++ {
			break
		}
		chapterSections = sections[startIndex:index]
	}

	// Get the slice of 'sections' that forms the backmatter
	backSections := sections[index:]

	// Struct to pass to the template
	data := struct {
		Title           string
		FrontSections   []SectionData
		HasParts        bool
		PartSections    []PartSectionData
		ChapterSections []SectionData
		BackSections    []SectionData
		Guides          []SectionData
	}{
		Title:           attributes["title"],
		FrontSections:   frontSections,
		HasParts:        hasParts,
		PartSections:    partSections,
		ChapterSections: chapterSections,
		BackSections:    backSections,
		Guides:          guides,
	}

	if err := tmpl.ExecuteTemplate(outfile, navTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genNCXFile generates the NCX file (for EPUB2 compatibility).
func genNCXFile() {
	fileName := "toc.ncx"
	fmt.Printf("Generating file %s (NCX) ... ", fileName)

	outfilespec := filepath.Join(packageDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		UUID     string
		Title    string
		Sections []SectionData
	}{
		UUID:     bookUUID,
		Title:    attributes["title"],
		Sections: sections,
	}

	if err := tmpl.ExecuteTemplate(outfile, ncxTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}

// genOPFFile generates the package (package.opf) file.
func genOPFFile() {
	fileName := "package.opf"
	fmt.Printf("Generating file %s (PACKAGE file) ... ", fileName)

	outfilespec := filepath.Join(packageDirSpec, fileName)
	outfile := fileutil.CreateFile(outfilespec)
	defer outfile.Close()

	_, hasISBN := attributes["isbn"]
	_, hasSeries := attributes["series"]
	_, hasRights := attributes["rights"]
	description := strings.Replace(attributes["description"], "<", "&lt;", -1)
	description = strings.Replace(description, ">", "&gt;", -1)

	// Struct to pass to the template
	data := struct {
		UUID        string
		HasISBN     bool
		ISBN        string
		Language    string
		Title       string
		TitleSort   string
		Author      string
		AuthorSort  string
		HasSeries   bool
		SeriesTitle string
		SeriesIndex string
		Publisher   string
		Description string
		Subjects    []string
		HasRights   bool
		Rights      string
		Created     string
		Modified    string
		CoverImage  ImageData
		Images      []ImageData
		Sections    []SectionData
		Guides      []SectionData
	}{
		UUID:        bookUUID,
		HasISBN:     hasISBN,
		ISBN:        attributes["isbn"],
		Language:    attributes["language"],
		Title:       attributes["title"],
		TitleSort:   attributes["title-sort"],
		Author:      attributes["author"],
		AuthorSort:  attributes["author-sort"],
		HasSeries:   hasSeries,
		SeriesTitle: attributes["series"],
		SeriesIndex: attributes["series-index"],
		Publisher:   attributes["publisher"],
		Description: description,
		Subjects:    strings.Split(attributes["subject"], ", "),
		HasRights:   hasRights,
		Rights:      attributes["rights"],
		Created:     attributes["created"],
		Modified:    attributes["modified"],
		CoverImage:  coverImage,
		Images:      images,
		Sections:    sections,
		Guides:      guides,
	}

	if err := tmpl.ExecuteTemplate(outfile, opfTemplate, data); err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	fmt.Println("done")
}
