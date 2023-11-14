package assistants

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type CreateMessageRequest struct {
	Role     string                 `json:"role"`
	Content  string                 `json:"content"`
	FileIDs  []string               `json:"file_ids"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ModifyMessageRequest struct {
	Metadata map[string]string `json:"metadata,omitempty"`
}

type DeleteMessageResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
	Object  string `json:"object"`
}

type ListMessagesResponse struct {
	Object  string    `json:"object"`
	Data    []Message `json:"data"`
	FirstID string    `json:"first_id"`
	LastID  string    `json:"last_id"`
	HasMore bool      `json:"has_more"`
}

func CreateMessage(client *http.Client, baseURL string, apiKey string, request CreateMessageRequest) (*Message, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/messages", bytes.NewBuffer(requestBody))
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

	var messageResponse Message
	if err := json.NewDecoder(resp.Body).Decode(&messageResponse); err != nil {
		return nil, err
	}

	return &messageResponse, nil
}

func ModifyMessage(client *http.Client, baseURL string, apiKey string, messageID string, request ModifyMessageRequest) (*Message, error) {

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", baseURL+"/messages/"+messageID, bytes.NewBuffer(requestBody))
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

	var messageResponse Message
	if err := json.NewDecoder(resp.Body).Decode(&messageResponse); err != nil {
		return nil, err
	}

	return &messageResponse, nil
}

func DeleteMessage(client *http.Client, baseURL string, apiKey string, messageID string) (*DeleteMessageResponse, error) {
	req, err := http.NewRequest("DELETE", baseURL+"/messages/"+messageID, nil)
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

	var deleteResponse DeleteMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResponse); err != nil {
		return nil, err
	}

	return &deleteResponse, nil
}

func ListMessages(client *http.Client, baseURL string, apiKey string, threadID string) (*ListMessagesResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/threads/"+threadID+"/messages", nil)
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

	var listResponse ListMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, err
	}

	return &listResponse, nil
}
