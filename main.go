package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var flagCollectionsLocation = flag.String("collection", "collection", "path to the collection (will be created if missing)")
var flagHymnBooks = flag.String("books", strings.Join(HymnBooks, ","), "hymn books to download (comma-delimited)")

func main() {
	flag.Parse()

	for _, hymnBook := range strings.Split(*flagHymnBooks, ",") {
		hymns, err := FetchHymns(hymnBook)
		if err != nil {
			log.Fatalf("failed to fetch hymns from %s: %s", hymnBook, err)
		}

		for _, hymn := range hymns.Data {
			if HymnExists(hymn.SongNumber, hymn.BookSlug, hymn.BookSectionTitle, hymn.Slug) {
				continue
			}
			log.Printf("downloading %s/%s-%s", hymnBook, hymn.SongNumber, hymn.Slug)
			hymnInfo, err := FetchHymn(hymn.Slug)
			if err == nil {
				saveHymn(hymn.SongNumber, hymn.BookSlug, hymn.BookSectionTitle, hymn.Slug, hymnInfo)
			}
		}
	}
}

func jsonPrint(obj interface{}) {
	stuff, err := json.MarshalIndent(obj, "", "  ")
	if err == nil {
		fmt.Printf("%s\n", stuff)
	}
}

func htmlToNewLines(what string) string {
	return stripHTMLTags(strings.ReplaceAll(what, "</p>", "\n"))
}

func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}
