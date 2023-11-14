package assistants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AssistantFileObject struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int64  `json:"created_at"`
	AssistantID string `json:"assistant_id"`
}

type CreateAssistantFileRequest struct {
	FileID string `json:"file_id"`
}

type DeleteAssistantFileResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
	Object  string `json:"object"`
}

type ListAssistantFilesResponse struct {
	Object  string                `json:"object"`
	Data    []AssistantFileObject `json:"data"`
	FirstID string                `json:"first_id"`
	LastID  string                `json:"last_id"`
	HasMore bool                  `json:"has_more"`
}

func ListAssistantFiles(client *http.Client, baseURL, apiKey, assistantID string, queryParams ...string) (*ListAssistantFilesResponse, error) {
	url := fmt.Sprintf("%s/assistants/%s/files%s", baseURL, assistantID, buildQueryString(queryParams...))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response ListAssistantFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
func CreateAssistantFile(client *http.Client, baseURL, apiKey, assistantID string, request CreateAssistantFileRequest) (*AssistantFileObject, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/assistants/%s/files", baseURL, assistantID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response AssistantFileObject
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
func GetAssistantFile(client *http.Client, baseURL, apiKey, assistantID, fileID string) (*AssistantFileObject, error) {
	url := fmt.Sprintf("%s/assistants/%s/files/%s", baseURL, assistantID, fileID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response AssistantFileObject
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func DeleteAssistantFile(client *http.Client, baseURL, apiKey, assistantID, fileID string) (*DeleteAssistantFileResponse, error) {
	url := fmt.Sprintf("%s/assistants/%s/files/%s", baseURL, assistantID, fileID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response DeleteAssistantFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
