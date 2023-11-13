package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// OpenAIError represents the standard error structure of OpenAI API responses
type OpenAIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param,omitempty"`
	} `json:"error"`
}

// Helper function to make an HTTP POST request
func doPostRequest(apiKey, url string, data interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("OpenAI-Beta", "assistants=v1")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if the response contains an error
	var apiError OpenAIError
	if err := json.Unmarshal(body, &apiError); err == nil {
		if apiError.Error.Message != "" {
			return nil, errors.New(apiError.Error.Message)
		}
	}

	return body, nil
}

// Helper function to make an HTTP GET request
func doGetRequest(apiKey, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if the response contains an error
	var apiError OpenAIError
	if err := json.Unmarshal(body, &apiError); err == nil {
		if apiError.Error.Message != "" {
			return nil, errors.New(apiError.Error.Message)
		}
	}

	return body, nil
}
