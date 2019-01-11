package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

const BaseUrl string = "https://api.imgur.com/3"

type JsonResponse map[string]interface{}

func imgurRequest(method string, action string) (*JsonResponse, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(method, BaseUrl+action, nil)
	if err != nil {
		return nil, err
	}

	// For our request we don't need an client id
	// clientId := os.Getenv("IMGUR-CLIENT-ID")
	// req.Header.Add("Authorization", "Client-ID "+clientId)

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	// fmt.Println(res.StatusCode)

	var data JsonResponse

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func imgurImageDownload(url string, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
