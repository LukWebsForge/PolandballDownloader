package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
)

const album string = "N63szEQ"

type ImageInfo struct {
	FileName    string   `json:"file_name"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	Creators    []string `json:"creators"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
}

func main() {

	/* if os.Getenv("IMGUR-CLIENT-ID") == "" {
		fmt.Println("Please set a imgur client id using the enviroment variable 'IMGUR-CLIENT-ID'")
		return
	} */

	res, err := imgurRequest("GET", "/album/"+album)
	if err != nil {
		fmt.Printf("Error while requesting the album: %v\n", err)
		return
	}

	data := (*res)["data"].(map[string]interface{})
	images := data["images"].([]interface{})

	fmt.Printf("Name: %v\n", data["title"])

	err = os.MkdirAll("images", os.ModePerm)
	if err != nil {
		fmt.Printf("Can't create image dir: %v\n", err)
		return
	}

	compile, err := regexp.Compile("/u/[A-Za-z0-9_-]+")
	if err != nil {
		fmt.Printf("Regex error: %v\n", err)
		return
	}

	dlChan := make(chan ImageInfo, 25)
	wg := sync.WaitGroup{}
	wg.Add(len(images))

	for i := 0; i < 50; i++ {
		go downloadPull(&wg, dlChan)
	}

	fmt.Print("Download: ")

	newData := make(map[int]ImageInfo)
	for i := 0; i < len(images); i++ {
		imageData := extractData(images[i].(map[string]interface{}), i, compile)
		newData[i] = imageData
		dlChan <- imageData
	}

	close(dlChan)
	wg.Wait()

	if writeNewData(newData) != nil {
		fmt.Printf("Error while writing data: %v\n", err)
		return
	}
}

func extractData(image map[string]interface{}, i int, compile *regexp.Regexp) ImageInfo {
	description := image["description"].(string)
	url := image["link"].(string)
	width := image["width"].(float64)
	height := image["height"].(float64)

	urlSplit := strings.Split(url, "/")
	filename := urlSplit[len(urlSplit)-1]

	creators := compile.FindAllString(description, -1)
	descShort := strings.Split(description, "\n\n")[0]

	if strings.Contains(descShort, "Created by") {
		descShort = ""
	}

	if i == 0 {
		descShort = "Cover"
	}

	return ImageInfo{
		FileName:    filename,
		Url:         url,
		Description: descShort,
		Creators:    creators,
		Width:       int(width),
		Height:      int(height),
	}
}

func downloadPull(wg *sync.WaitGroup, dlChan chan ImageInfo) {
	for dlReq := range dlChan {
		fmt.Print("|")
		err := imgurImageDownload(dlReq.Url, "images/"+dlReq.FileName)
		wg.Done()
		if err != nil {
			fmt.Printf("Download error: %v\n", err)
		}
	}
}

func writeNewData(data map[int]ImageInfo) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("data.json", bytes, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
