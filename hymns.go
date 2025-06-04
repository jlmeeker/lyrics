package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var HymnBooks = []string{"hymns-for-home-and-church", "hymns", "childrens-songbook"}
var HymnBookBaseURL = "https://www.churchofjesuschrist.org/media/music/api?type=songsFilteredList&lang=eng&batchSize=500&identifier="
var HymnBookQueryString = `{"lang":"eng","limit":500,"offset":0,"orderByKey":["bookSongPosition"],"bookQueryList":["BOOK"]}`
var HymnBaseURL = "https://www.churchofjesuschrist.org/media/music/songs/SLUG?lang=en"

type HymnList struct {
	Data []Hymn
}

type Hymn struct {
	Title               string
	Slug                string
	BookSlug            string
	BookSectionTitle    string
	SongNumber          string
	SheetMusicAvailable bool
}
type Verse struct {
	Body        string `json:"body"`
	VerseNumber int    `json:"verseNumber"`
	URL         string `json:"url"`
}

type HymnInfo struct {
	Data struct {
		SlugName string `json:"slugName"`
		SongData struct {
			Assets []struct {
				MediaObject struct {
					Description string `json:"description"`
				} `json:"mediaObject"`
			} `json:"assets"`
			Lyrics struct {
				VerseList []struct {
					Verse Verse `json:"verse"`
				} `json:"verseList"`
			} `json:"lyrics"`
		} `json:"songData"`
	} `json:"data"`
}

func HymnExists(number, book, section, slug string) bool {
	var destDir = filepath.Join("collection", book)
	var destSongDir = filepath.Join(destDir, "songs")
	var lyricFileName = number + "-" + slug + ".txt"
	var lyricFilePath = filepath.Join(destSongDir, lyricFileName)
	_, err := os.Stat(lyricFilePath)
	return err == nil
}

func FetchHymns(book string) (HymnList, error) {
	var hymns HymnList
	queryString := strings.ReplaceAll(HymnBookQueryString, "BOOK", book)
	res, err := http.Get(HymnBookBaseURL + url.QueryEscape(queryString))
	if err != nil {
		return hymns, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return hymns, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return hymns, err
	}
	err = json.Unmarshal(body, &hymns)
	return hymns, err
}

func FetchHymn(slug string) (HymnInfo, error) {
	var hymnInfo HymnInfo
	res, err := http.Get(strings.ReplaceAll(HymnBaseURL, "SLUG", slug))
	if err != nil {
		return hymnInfo, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return hymnInfo, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return hymnInfo, err
	}

	selection := doc.Find("script")
	selection.Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "verseNumber") {
			jsonString := strings.TrimPrefix(scriptText, "window.renderData=")
			json.Unmarshal([]byte(jsonString), &hymnInfo)
			s.End()
		}
	})
	return hymnInfo, err
}

func saveHymn(number, book, section, slug string, hymnInfo HymnInfo) error {
	var destDir = filepath.Join("collection", book)
	var err error

	// save lyric file (ignore if exists)
	var destSongDir = filepath.Join(destDir, "songs")
	if err = os.MkdirAll(destSongDir, 0755); err != nil {
		return err
	}
	var lyricFileName = number + "-" + slug + ".txt"
	var lyricFilePath = filepath.Join(destSongDir, lyricFileName)
	if _, err = os.Stat(lyricFilePath); err != nil {
		var data []byte
		data = append(data, []byte(hymnVersesAsString(hymnInfo))...)
		err = os.WriteFile(lyricFilePath, data, 0644)
	}

	// create section symlink (ignore if exists)
	var destSectionDir = filepath.Join(destDir, section)
	if err = os.MkdirAll(destSectionDir, 0755); err != nil {
		return err
	}
	var sectionLinkFilePath = filepath.Join(destSectionDir, number+"-"+slug+".txt")
	if _, err = os.Stat(sectionLinkFilePath); err != nil {
		var data []byte
		data = append(data, []byte(hymnVersesAsString(hymnInfo))...)
		err = os.Symlink(filepath.Join("..", "songs", lyricFileName), sectionLinkFilePath)
	}

	return err
}

func hymnVersesAsString(hymnInfo HymnInfo) string {
	var versesString string
	for _, verse := range hymnInfo.Data.SongData.Lyrics.VerseList {
		if verse.Verse.VerseNumber == 0 {
			continue
		}
		versesString += fmt.Sprintf("%s\n", htmlToNewLines(verse.Verse.Body))
	}
	return versesString
}

func printVerse(verse Verse) {
	fmt.Printf("%s\n", htmlToNewLines(verse.Body))
}

func PrintVerses(hymnInfo HymnInfo) {
	fmt.Printf("Title: %s\n", hymnInfo.Data.SongData.Assets[0].MediaObject.Description)
	fmt.Println("")
	for _, verse := range hymnInfo.Data.SongData.Lyrics.VerseList {
		if verse.Verse.VerseNumber == 0 {
			continue
		}
		printVerse(verse.Verse)
	}
}
