package assistants

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type CreateThreadRequest struct {
	Messages []CreateMessageRequest `json:"messages"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ModifyThreadRequest struct {
	Metadata map[string]interface{} `json:"metadata"`
}

type DeleteThreadResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
	Object  string `json:"object"`
}

type ListThreadsResponse struct {
	Object  string   `json:"object"`
	Data    []Thread `json:"data"`
	FirstID string   `json:"first_id"`
	LastID  string   `json:"last_id"`
	HasMore bool     `json:"has_more"`
}

func CreateThread(client *http.Client, baseURL string, apiKey string, request CreateThreadRequest) (*Thread, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/threads", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var threadResponse Thread
	if err := json.NewDecoder(resp.Body).Decode(&threadResponse); err != nil {
		return nil, err
	}

	return &threadResponse, nil
}

func ModifyThread(client *http.Client, baseURL string, apiKey string, threadID string, request ModifyThreadRequest) (*Thread, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", baseURL+"/threads/"+threadID, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var threadResponse Thread
	if err := json.NewDecoder(resp.Body).Decode(&threadResponse); err != nil {
		return nil, err
	}

	return &threadResponse, nil
}

func DeleteThread(client *http.Client, baseURL string, apiKey string, threadID string) (*DeleteThreadResponse, error) {
	req, err := http.NewRequest("DELETE", baseURL+"/threads/"+threadID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var deleteResponse DeleteThreadResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResponse); err != nil {
		return nil, err
	}

	return &deleteResponse, nil
}

func ListThreads(client *http.Client, baseURL string, apiKey string) (*ListThreadsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/threads", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var listResponse ListThreadsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, err
	}

	return &listResponse, nil
}
