# Creates, checks and packages EPUB3 books
#
Function Usage {
  Write-Host "EPUBGen: Creates, checks and packages EPUB3 books"
  Write-Host "Usage:"
  Write-Host "  epg <bookname>"
}

if ($args.length -eq 0) {
  Usage
  exit
}

$epubcheck = "C:\Java\epubcheck-4.2.6\epubcheck.jar"
$book = $args[0]

Write-Host "Generating $book..."
./epubgen $book
if ($?) {
  Write-Host ""
  Write-Host "Checking $book..."
  java -jar $epubcheck data\generated\$book --profile default -v 3.0 --mode exp -save
}
