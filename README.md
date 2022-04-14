<div id="banner" align="center">
    <br />
    <h1>EP3Gen</h1>
    <h3>Free/Libre Open Source e-book generator written in Go</h3>
</div>

# Overview
**EP3Gen** is my attempt at learning the Go programming language. I decided to build a command line program to help generate and package an EPUB3-compliant e-book. Instead of building a fancy visual editor to create an e-book, why not use a single HTML file as the source for the e-book? As a developer, I have access to many powerful free and open source IDEs and text editors such as Eclipse, ItelliJ IDEA Community and VSCodium/VSCode with which to create the HTML file. If you prefer the fancy visual editor take a look at [Sigil.](https://sigil-ebook.com/)

**EP3Gen** is a text processor which reads and processes the HTML source file and generates all the files in a directory structure expected by EPUB3. The generated directory and its subdirectories and contents can then be checked and packaged into an `.epub` file by running the utility [EPUBCheck.](https://github.com/w3c/epubcheck/releases/)

# Getting Started
To start using EP3Gen to generate your own EPUB3 e-books, at a minimum download the following:

1. `ep3gen.exe` if you are using Windows. If you using either Linux or MacOS, the easiest way is to clone this repo and build the executable as follows: `go build -o ep3gen *.go`. Make sure to make the file executable.

1. The `data` directory and all of its contents.

# Generate the sample e-book
Under the `data/source` directory you can find the folder `rls-treasure-island`. This folder contains 3 files: `author.jpeg`, `cover.jpeg` and `source.html`. The last one contains the complete source of the book *Treasure Island* by Robert Louis Stevenson which is in the Public Domain.

Issue the following command under Powershell, Command Prompt, Git-Bash or Terminal:

    ep3gen rls-treasure-island    // use ./ep3gen under Linux or MacOS

Under the `data/generated` directory, you can see the folder `rls-treasure-island` which contains the expanded EPUB3 e-book package.

To actually turn it into an `.epub` file, you need the [EPUBCheck](https://github.com/w3c/epubcheck/releases/) utility. Since it is a Java JAR file, you need the Java runtime installed on your system before you can use it.

Once you have it, run the following command to check the integrity of the e-book and at the same time package it into an `.epub` file, which is actually a ZIP archive:

    java -jar [PATH]/epubcheck.jar \
      data/generated/rls-treasure-island \
      --profile default -v 3.0 --mode exp -save

You will see the following output:

    Validating using EPUB version 3.2 rules.
    No errors or warnings detected.
    Messages: 0 fatals / 0 errors / 0 warnings / 0 infos

    EPUBCheck completed

I have created a Windows Powershell script `epg.ps1` to automate the two steps above. You can customize it to your liking or use it as the basis for your own Linux or MacOS bash script.

Under the `data/generated` directory, you can see the file `rls-treasure-island.epub`. 

Use your favorite e-book reader to read the new *Treasure Island* e-book. I use Calibre e-book viewer on Windows. You can also transfer the book to your phone and read it there.

# Create you own EPUB3 e-book
Here's how you can create your own EPUB3-compliant e-book:

1. Create a new folder under the `data/source` directory for the book.

2. You need at least two files: the cover image file and the source HTML file.

3. The cover image file can be any name but must have either `.png` or `.jpeg` extension. It is easy to put in support for `.jpg` extension but it is even easier just to rename it to `.jpeg`.

4. The HTML source file must be named `source.html`. It should be a valid HTML5 file.

5. I make use of the `<meta>` elements under the `<head>` element for specifying the attributes used by EPUB3 e-books.

6. I also use specific HTML comments as directives to organize the various sections (such as cover page, copyright section, preface, parts, chapter, appendices, etc).

7. See the sample file in `data/source/rls-treasure-island/source.html` to see how the HTML file is contructed. 

# Overriding default locations
You can override the default locations by editing the file `config.yaml` which should be in the current directory whenever you run the commands. The default `config.yaml` is:

    # The parent directory of all e-book source directories
    source_dir: ./data/source

    # The parent directory of all e-book generated contents
    target_dir: ./data/generated

    # Where you can find the static files and the CSS file
    resource_dir: ./data/etc

    # Where you can find the Go text/template source files
    templates_dir: ./data/templates

# Attributes
Attributes are specified as `<meta>` elements under the `<head>` element of the HTML file. It has the format:

    <meta name="attribute" content="value"/>

The following attributes are mandatory:

1. `version`: It must have the value of `epub3`.

1. `title`: It should contain the name of the book as displayed on the cover page.

1. `title-sort`: It should contain the name of the book for use in a sorted list useful for searching.

1. `author`: It should contain the name of the main author as displayed on the cover page.

1. `author-sort`: It should contain the name of the main author for use in a sorted list useful for searching.

1. `published`: It should contain the release month and year or at least the year of publication, such as "April 2022".

1. `publisher`: It should contain the name of the publisher such as your company name, “Unknown”, etc.

1. `language`: It should contain the standard code for a language, such as `en` or `en-US`.

1. `cover-image`: should contain the name of the image file for the cover page, usually `cover.png` or `cover.jpeg`.

The following attributes are optional:

1. `author2`: It should contain the name of the second author as displayed on the cover page, if any.

1. `author3`: It should contain the name of the third author as displayed on the cover page, if any.

1. `series`: It should contain the name of the series for which this book is a part of as displayed on the cover page, if any.

1. `series-index`: It should contain the volume number (1,2,3,...) of this book in the series, such as, “Volume 1 of the XXX trilogy”.

1. `images`: It should contain the comma-separated list of image files for all images used in the book other than the cover image, such as, `image1.png,image2.png`. Make sure there are no spaces in the list.

1. `titlepage`: This is optional but if given must contain one of the following: 1) `default`: EP3Gen will generate a default title page for the book; 2) `anyname.png` or `anyname.jpeg`: EP3Gen will use the image file specified as the title page, and it must be one of the files listed above; 3) `custom`: You will need to specify a `<!--titlepage-->` directive with one or more custom HTML lines to use as the title page. If no `titlepage` attribute is given, it is the same as specifying `default`.

1. `description`: It should be used to describe the book for marketing purposes which the user can read before opening the book for reading.

1. `subject`: A comma-separated list of subjects describing the various classifications of the book such as *General, Fiction, Action &amp; Adventure*

1. `created`: The date and time the book was first created in the RFC3339 format. If not supplied, EP3Gen will use the current date and time as the creation date. EP3Gen will automatically add the `modified` attribute in any case which is required by the EPUB3 specifications.

1. `subtitle`: It should contain the subtitle of the book as displayed on the book cover and title page, if available.

1. `isbn`: If you have the ISBN for the book, you can specify it here.

1. `rights`: A short copyright statement if available, such as “Copyright © 2022 by Roslan Amir. All rights reserved.”

# Directives
Directives are specified as HTML comments inserted among the `<hx>`, `<p>`, etc elements and control the organization of the book into multiple sections, parts and chapters, etc. The following directives are mandatory:

1. `<!--end-->`: Must be the last directive in the HTML file, usually put just before the closing `</body>` element.

1. `<!--titlepage-->`: This is mandatory if you specify the attribute `titlepage` as `custom`. It must be the first directive after the `<body>` element. It should contain one or more formatted HTML elements and will be used to display the title page.

1. `<!--copyright-->`: This is mandatory and must be present. The first line must be `<h1>&#160;</h1>` to indicate en empty heading for this section. Must be followed by one of more formatted HTML to display the copyright section of the book. The section heading is hard-coded as `Copyright` for display in the TOC.

1. `<!--chapter-->`: At least one of this must be present in the source HTML file. This represents a chapter or section in the book. The first line must contain the chapter heading with one of the `<h1>`, `<h2>` or `<h3>` elements. It must be followed by one or more formatted HTML elements.

In large books, the chapters may be broken up into multiple parts. In this case, you may put a `<!--part-->` directive before a group of chapters:

1. `<!--part-->`: The first line must contain the part heading with one of the `<h1>`, `<h2>` or `<h3>` elements. It must be followed by zero or more formatted HTML elements.

The following directives are optional. The first line must consist of the header HTML element with one of the `<h1>`, `<h2>` or `<h3>` to be used as the section heading. If the section heading is not applicable, use `<h1>&#160;</h1>` for the first line and a default heading will be used in the TOC:

1. `<!--bibliography-->`: May occur at most once at the front part of the book. Usually used to list the other books by the same author(s).

1. `<!--acknowledgments-->`: May occur at most once at the front part of the book. Usually used by the author to acknowledge the contributions of other people to the book.

1. `<!--dedication-->`: May occur at most once at the front part of the book. Usually used by the author to dedicate the book to someone else.

1. `<!--epigraph-->`: May occur at most once at the front part of the book.

1. `<!--foreword-->`: May occur at most once at the front part of the book.

1. `<!--introduction-->`: May occur at most once at the front part of the book.

1. `<!--prologue-->`: May occur at most once at the front part of the book.

1. `<!--preamble-->`: May occur multiple times. Acts as the generic section for the front part of the book.

1. `<!--afterword-->`: May occur at most once at the back part of the book.

1. `<!--epilogue-->`: May occur at most once at the back part of the book.

1. `<!--appendix-->`: May occur multiple times. Acts as the generic section for the back part of the book.

# Stylesheet
Under the `data/etc` folder you can find the minimal `stylesheet.css` file for formatting the HTML elements used the book. Feel free to modify it to your heart's content. Make sure it is named `stylesheet.css`.

# Contributing
Please read our [Contributing Guide](https://github.com/roslamir/ep3gen/blob/main/CONTRIBUTING.md) before submitting a pull request to the project.

# License
Licensed under the [MIT License](https://github.com/roslamir/ep3gen/blob/main/LICENSE.md)
