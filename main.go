package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type image struct {
	LangName  string
	LangColor string
	TextColor string
}

type imageLink struct {
	Name        string
	EncodedName string
	ImageName   string
	HexColor    string
}

type readme []imageLink

func main() {

	// parse templates
	tmpl, err := template.ParseFiles("./image.tmpl", "./readme.tmpl")
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

	// ensure svgs directory exists and is empty
	err = os.RemoveAll("./svgs")
	if err != nil {
		if err != os.ErrNotExist {
			log.Fatal(err)
		}
	}

	err = os.Mkdir("./svgs", 0755)
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

				// language name as base64 raw url encoding
				base64Name := base64.RawURLEncoding.EncodeToString([]byte(lang))

				// timestamp for svg filenames so they aren't cached
				unixTime := time.Now().Unix()

				// create the svg image file name
				imageName := fmt.Sprintf("%s-%d.svg", base64Name, unixTime)

				svgBuffer := bytes.Buffer{}

				img := image{
					LangColor: color,
				}

				err = tmpl.ExecuteTemplate(&svgBuffer, "image.tmpl", img)
				if err != nil {
					log.Fatal(err)
				}

				err = ioutil.WriteFile("./svgs/"+imageName, svgBuffer.Bytes(), 0644)
				if err != nil {
					log.Fatal(err)
				}

				// encode any spaces
				encodedName := strings.Replace(lang, " ", "%20", -1)

				// encode any single quotes
				encodedName = strings.Replace(encodedName, "'", "&apos;", -1)

				// add language to readme
				readmeData = append(readmeData, imageLink{
					Name:        lang,
					EncodedName: encodedName,
					ImageName:   imageName,
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
