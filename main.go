package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type language struct {
	Name        string
	EncodedName string
	HexColor    string
}

type readme []language

func main() {

	// parse templates
	tmpl, err := template.ParseFiles("./readme.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	// get and decode yaml
	baseURL := "https://raw.githubusercontent.com"
	req, err := http.Get(baseURL + "/github/linguist/master/lib/linguist/languages.yml")
	if err != nil {
		log.Fatal(err)
	}

	defer req.Body.Close()

	var languageData map[string]map[string]interface{}
	err = yaml.NewDecoder(req.Body).Decode(&languageData)
	if err != nil {
		log.Fatal(err)
	}

	// generate svgs and readme
	readmeData := readme{}

	// sort languages by name
	keys := make([]string, len(languageData))
	for key := range languageData {
		keys = append(keys, key)
	}
	less := func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	}
	sort.Slice(keys, less)

	// find languages with colors
	for _, lang := range keys {
		meta := languageData[lang]
		if meta["color"] != nil {
			color, ok := meta["color"].(string)
			if ok {

				// encode any spaces
				encodedName := strings.Replace(lang, " ", "%20", -1)

				// encode any single quotes
				encodedName = strings.Replace(encodedName, "'", "&apos;", -1)

				// add language to readme
				readmeData = append(readmeData, language{
					Name:        lang,
					EncodedName: encodedName,
					HexColor:    color,
				})
			}
		}
	}

	readmeBuffer := bytes.Buffer{}
	err = tmpl.ExecuteTemplate(&readmeBuffer, "readme.tmpl", readmeData)
	if err != nil {
		log.Fatal(err)
	}

	// create README.md
	err = ioutil.WriteFile("./README.md", readmeBuffer.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

}
