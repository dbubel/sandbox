package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	gourl "net/url"
	"os"
	"path"

	"github.com/PuerkitoBio/purell"
)

func main() {
	// Open the CSV file
	file, err := os.Open("data.csv")
	if err != nil {
		log.Fatalf("Failed to open CSV file: %s", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all records from the CSV
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %s", err)
	}

	// Check if the CSV has the expected header
	if len(records) == 0 || len(records[0]) < 3 {
		log.Fatalf("CSV file does not have the expected format")
	}

	// Iterate over the records and generate the SQL statements
	for _, record := range records[1:] { // Skip the header
		url := record[1]
		url, err = NormURL(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		linkType := record[2]
		statement := fmt.Sprintf("UPDATE content SET link_type = '%s' WHERE normalized_url = '%s';", linkType, url)
		fmt.Println(statement)
	}
}

func NormURL(url string) (string, error) {
	// This lib gets us _most_ of the way there
	url, err := purell.NormalizeURLString(url,
		purell.FlagLowercaseHost| // http://HOST -> http://host
			purell.FlagUppercaseEscapes| //
			purell.FlagDecodeUnnecessaryEscapes| // http://host/t%41 -> http://host/tA
			purell.FlagRemoveDefaultPort| // http://host:80 -> http://host
			purell.FlagRemoveTrailingSlash| // http://host/path/ -> http://host/path
			purell.FlagRemoveDotSegments| // http://host/path/./a/b/../c -> http://host/path/a/c
			purell.FlagRemoveDirectoryIndex| // http://host/path/index.html -> http://host/path/
			purell.FlagRemoveFragment| // http://host/path#fragment -> http://host/path
			purell.FlagRemoveDuplicateSlashes| // http://host/path//a///b -> http://host/path/a/b
			purell.FlagRemoveWWW) // http://www.host -> http://host
	if err != nil {
		return url, err
	}
	// Load our intermediate result into a url obj, strip the protocol and the query string
	u, err := gourl.Parse(url)
	if err != nil {
		return url, err
	}

	if path.Join(u.Host, u.Path) == "" {
		return "", errors.New("invalid url")
	}
	// RescoreExtracted and return just the parts we are interested in
	return path.Join(u.Host, u.Path), nil
}
