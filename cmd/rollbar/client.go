package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
)

type RollbarResponse struct {
	Err    int `json:"err"`
	Result struct {
		Instances []Instance `json:"instances"`
	} `json:"result"`
}

func ListOccurrences() {
	accessToken := viper.GetString("ROLLBAR_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal().Msg("Missing Rollbar access token")
	}

	url := "https://api.rollbar.com/api/1/instances"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating request")
	}

	req.Header.Add("X-Rollbar-Access-Token", accessToken)
	req.Header.Add("Accept", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Error sending request")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// check response status code
	// read body into bytes

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading response body")
	}

	if resp.StatusCode != 200 {
		log.Fatal().
			Str("body", string(body)).
			Msg("Error response from Rollbar API")
	}

	var response RollbarResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println(string(body))
		log.Fatal().Err(err).
			Msg("Error decoding response")
	}

	for _, instance := range response.Result.Instances {
		fmt.Println(instance)
	}
}
