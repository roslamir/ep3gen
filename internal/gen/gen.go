// Copyright (C) 2022-2023, Roslan Amir. All rights reserved.
// Created on: 13-Apr-2022
//
// Functions to generate the various sections of the e-book

package gen

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/roslamir/ep3gen/internal/fileutil"
	"github.com/roslamir/ep3gen/internal/parm"
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
	FileName  string // image file name with extension
	MediaType string // the media type (png/jpeg) based on extension
}

var (
	tmpl           *template.Template
	sourceDirSpec  string // the full path for the source directory
	targetDirSpec  string // the full path for the output directory
	packageDirSpec string // the full path for the OEBPS directory
	textDirSpec    string // the full path for the OEBPS/Text directory
)

// LoadTemplates loads in the template files and panics if any error occurs.
func LoadTemplates() {
	tmpl = template.Must(template.ParseFiles(
		filepath.Join(parm.TemplatesDir, coverTemplate),
		filepath.Join(parm.TemplatesDir, defaultTitlepageTemplate),
		filepath.Join(parm.TemplatesDir, imageTitlepageTemplate),
		filepath.Join(parm.TemplatesDir, frontmatterTemplate),
		filepath.Join(parm.TemplatesDir, bodymatterTemplate),
		filepath.Join(parm.TemplatesDir, backmatterTemplate),
		filepath.Join(parm.TemplatesDir, navTemplate),
		filepath.Join(parm.TemplatesDir, ncxTemplate),
		filepath.Join(parm.TemplatesDir, opfTemplate),
	))
}

// Init creates the EPUB directory tree.
func Init(sourceDir, targetDir string) {
	sourceDirSpec = sourceDir
	targetDirSpec = targetDir
	packageDirSpec = filepath.Join(targetDirSpec, "OEBPS")
	textDirSpec = filepath.Join(packageDirSpec, "Text")
}

// GenCoverSection generates the cover page section.
func (b *InputBuffer) GenCoverSection() {
	section := SectionData{
		ID:       "cover",
		EpubType: "cover",
		Heading:  "Cover Page",
	}
	b.sections = append(b.sections, section)
	b.guides = append(b.guides, section)

	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		CoverImage ImageData
		Title      string
	}{
		CoverImage: b.coverImage,
		Title:      b.attributes["title"],
	}

	if err := tmpl.ExecuteTemplate(outfile, coverTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenTitlePageSection generates the title page section.
// If the attribute "titlepage" is not given or has the value of "default", we generate the default title page section.
// If it has the value of "custom", the first directive encountered must be "<!--titlepage-->" and it must be followed
// by one or more formatted HTML lines making up the title page section.
// Any other value is assumed to be the name of an image file with either "png" or "jpeg" extension which will be used
// as the title page.
func (b *InputBuffer) GenTitlePageSection() {
	var titlePage string
	if titlePage = b.attributes["titlepage"]; titlePage == "" {
		titlePage = "default"
	}

	section := SectionData{
		ID:       "titlepage",
		EpubType: "titlepage",
		Heading:  "Title Page",
	}
	b.sections = append(b.sections, section)
	b.guides = append(b.guides, section)

	switch titlePage {
	case "default":
		b.GenDefaultTitlePageSection(section)

	case "custom":
		b.NextLine()
		if b.CurrLine == "<!--titlepage-->" {
			b.NextLine()
			b.GenFrontMatterSection(section)
		} else {
			panic("epubgen: <!--titlepage--> directive expected")
		}

	default: // assumes titlepage contains an image file name to be used for the title page
		parts := strings.Split(titlePage, ".")
		mediaType := parts[1]
		if mediaType != "png" && mediaType != "jpeg" {
			panic("epubgen: only image files with extension 'png' or 'jpeg' are accepted")
		}
		image := ImageData{
			FileName:  titlePage,
			MediaType: mediaType,
		}
		b.GenImageTitlePageSection(section, image)
	}
}

// GenDefaultTitlePageSection generates the default title page section.
func (b *InputBuffer) GenDefaultTitlePageSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	var err error
	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Struct to pass to the template
	_, hasSubtitle := b.attributes["subtitle"]
	_, hasSeries := b.attributes["series"]
	_, hasAuthor2 := b.attributes["author2"]
	_, hasAuthor3 := b.attributes["author3"]
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
		Title:       b.attributes["title"],
		HasSubtitle: hasSubtitle,
		Subtitle:    b.attributes["subtitle"],
		HasSeries:   hasSeries,
		Series:      b.attributes["series"],
		SeriesIndex: b.attributes["series-index"],
		Author:      b.attributes["author"],
		HasAuthor2:  hasAuthor2,
		Author2:     b.attributes["author2"],
		HasAuthor3:  hasAuthor3,
		Author3:     b.attributes["author3"],
		Publisher:   b.attributes["publisher"],
		Published:   b.attributes["published"],
	}

	if err = tmpl.ExecuteTemplate(outfile, defaultTitlepageTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenImageTitlePageSection generates the title page section comprising a single image.
func (b *InputBuffer) GenImageTitlePageSection(section SectionData, image ImageData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Image    ImageData
		Heading  string
	}{
		Title:    b.attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Image:    image,
		Heading:  section.Heading,
	}

	if err := tmpl.ExecuteTemplate(outfile, imageTitlepageTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenCopyrightSection generates the mandatory copyright section file.
// On entry, currLine should contain the directive <!--copyright-->.
func (b *InputBuffer) GenCopyrightSection(currDate string) {
	if b.CurrLine != "<!--copyright-->" {
		panic("epubgen: <!--copyright--> directive expected")
	}
	b.NextLine()

	section := SectionData{
		ID:       "copyright",
		EpubType: "copyright-page",
		Heading:  "Copyright",
	}
	b.sections = append(b.sections, section)

	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0, 50)
	for {
		sectionLines = append(sectionLines, b.CurrLine)
		b.NextLine()
		if strings.HasPrefix(b.CurrLine, "<!--") {
			break
		}
	}

	// Append the e-book generation date to the end of the copyright section.
	sectionLines = append(sectionLines, `<p class="copy">&#160;</p>`)
	sectionLines = append(sectionLines, `<p class="copy">This e-book generated on `+currDate+`</p>`)

	// Struct to pass to the template
	data := struct {
		Title    string
		ID       string
		EpubType string
		Lines    []string
	}{
		Title:    b.attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, frontmatterTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenFrontMatterSection generates one of the various frontmatter sections file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func (b *InputBuffer) GenFrontMatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0, 50)
	for {
		sectionLines = append(sectionLines, b.CurrLine)
		b.NextLine()
		if strings.HasPrefix(b.CurrLine, "<!--") {
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
		Title:    b.attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, frontmatterTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenBodyMatterSection generates the bodymatter (part or chapter) section file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func (b *InputBuffer) GenBodyMatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0, 50)
	for {
		sectionLines = append(sectionLines, b.CurrLine)
		b.NextLine()
		if strings.HasPrefix(b.CurrLine, "<!--") {
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
		Title:    b.attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, bodymatterTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenBackMatterSection generates the copyright section file.
// On entry, currLine contains the first line of this section, either <h1>, <h2> or <h3> tag.
func (b *InputBuffer) GenBackMatterSection(section SectionData) {
	fileName := section.ID + ".xhtml"
	fmt.Printf("Generating file %s (%s) ... ", fileName, section.Heading)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Read in the lines making up the section and stop when another directive line is encountered.
	sectionLines := make([]string, 0, 50)
	for {
		sectionLines = append(sectionLines, b.CurrLine)
		b.NextLine()
		if strings.HasPrefix(b.CurrLine, "<!--") {
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
		Title:    b.attributes["title"],
		ID:       section.ID,
		EpubType: section.EpubType,
		Lines:    sectionLines,
	}
	if err := tmpl.ExecuteTemplate(outfile, backmatterTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// PartSectionData holds the list of part sections with their chapter sections.
type PartSectionData struct {
	Part     SectionData
	Chapters []SectionData
}

// GenNAVFile generates the NAV (TOC) file (required for EPUB3).
func (b *InputBuffer) GenNAVFile() {
	fileName := "nav.xhtml"
	fmt.Printf("Generating file %s (TOC) ... ", fileName)

	outfile := fileutil.CreateFile(filepath.Join(textDirSpec, fileName))
	defer outfile.Close()

	// Get the slice of 'sections' that forms the frontmatter
	var index int
	var section SectionData
	for index, section = range b.sections {
		if section.EpubType == "part" || section.EpubType == "chapter" {
			break
		}
	}
	frontSections := b.sections[:index]

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
			section = b.sections[index]
			if section.EpubType == "part" {
				if firstTime {
					firstTime = false
					partSections = make([]PartSectionData, 0, 10)
					currPart = section
					startIndex = index + 1
				} else {
					partSection := PartSectionData{
						Part:     currPart,
						Chapters: b.sections[startIndex:index],
					}
					partSections = append(partSections, partSection)
					currPart = section
					startIndex = index + 1
				}
			} else if section.EpubType != "chapter" {
				partSection := PartSectionData{
					Part:     currPart,
					Chapters: b.sections[startIndex:index],
				}
				partSections = append(partSections, partSection)
				break
			}
			index++
		}
	} else {
		// Get the slice of 'sections' that forms the chapters (no parts)
		startIndex := index
		for ; b.sections[index].EpubType != "chapter"; index++ {
			break
		}
		chapterSections = b.sections[startIndex:index]
	}

	// Get the slice of 'sections' that forms the backmatter
	backSections := b.sections[index:]

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
		Title:           b.attributes["title"],
		FrontSections:   frontSections,
		HasParts:        hasParts,
		PartSections:    partSections,
		ChapterSections: chapterSections,
		BackSections:    backSections,
		Guides:          b.guides,
	}

	if err := tmpl.ExecuteTemplate(outfile, navTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenNCXFile generates the NCX file (for EPUB2 compatibility).
func (b *InputBuffer) GenNCXFile() {
	fileName := "toc.ncx"
	fmt.Printf("Generating file %s (NCX) ... ", fileName)

	outfile := fileutil.CreateFile(filepath.Join(packageDirSpec, fileName))
	defer outfile.Close()

	// Struct to pass to the template
	data := struct {
		UUID     string
		Title    string
		Sections []SectionData
	}{
		UUID:     parm.BookUUID,
		Title:    b.attributes["title"],
		Sections: b.sections,
	}

	if err := tmpl.ExecuteTemplate(outfile, ncxTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// GenOPFFile generates the package file (package.opf).
func (b *InputBuffer) GenOPFFile() {
	fileName := "package.opf"
	fmt.Printf("Generating file %s (PACKAGE file) ... ", fileName)

	outfile := fileutil.CreateFile(filepath.Join(packageDirSpec, fileName))
	defer outfile.Close()

	_, hasISBN := b.attributes["isbn"]
	_, hasSeries := b.attributes["series"]
	_, hasRights := b.attributes["rights"]
	description := strings.Replace(b.attributes["description"], "<", "&lt;", -1)
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
		UUID:        parm.BookUUID,
		HasISBN:     hasISBN,
		ISBN:        b.attributes["isbn"],
		Language:    b.attributes["language"],
		Title:       b.attributes["title"],
		TitleSort:   b.attributes["title-sort"],
		Author:      b.attributes["author"],
		AuthorSort:  b.attributes["author-sort"],
		HasSeries:   hasSeries,
		SeriesTitle: b.attributes["series"],
		SeriesIndex: b.attributes["series-index"],
		Publisher:   b.attributes["publisher"],
		Description: description,
		Subjects:    strings.Split(b.attributes["subject"], ", "),
		HasRights:   hasRights,
		Rights:      b.attributes["rights"],
		Created:     b.attributes["created"],
		Modified:    b.attributes["modified"],
		CoverImage:  b.coverImage,
		Images:      b.images,
		Sections:    b.sections,
		Guides:      b.guides,
	}

	if err := tmpl.ExecuteTemplate(outfile, opfTemplate, data); err != nil {
		panic(err)
	}

	fmt.Println("done")
}

// CopyStaticFiles copies	the control files, the stylesheet and the image files.
func (b *InputBuffer) CopyStaticFiles() {
	// <targetdir>/mimetype
	sourceFileSpec := filepath.Join(parm.ResourceDir, "mimetype")
	targetFileSpec := filepath.Join(targetDirSpec, "mimetype")
	fileutil.CopyFile(sourceFileSpec, targetFileSpec)

	// <targetdir>/META-INF/container.xml
	sourceFileSpec = filepath.Join(parm.ResourceDir, "container.xml")
	targetFileSpec = filepath.Join(targetDirSpec, "META-INF", "container.xml")
	fileutil.CopyFile(sourceFileSpec, targetFileSpec)

	// <targetdir>/OEBPS/Styles/stylesheet.css
	sourceFileSpec = filepath.Join(parm.ResourceDir, "stylesheet.css")
	targetFileSpec = filepath.Join(packageDirSpec, "Styles", "stylesheet.css")
	fileutil.CopyFile(sourceFileSpec, targetFileSpec)

	// <targetdir>/OEBPS/Images/*
	sourceFileSpec = filepath.Join(sourceDirSpec, b.coverImage.FileName)
	targetFileSpec = filepath.Join(packageDirSpec, "Images", b.coverImage.FileName)
	fileutil.CopyFile(sourceFileSpec, targetFileSpec)

	for _, image := range b.images {
		sourceFileSpec = filepath.Join(sourceDirSpec, image.FileName)
		targetFileSpec = filepath.Join(packageDirSpec, "Images", image.FileName)
		fileutil.CopyFile(sourceFileSpec, targetFileSpec)
	}
}
