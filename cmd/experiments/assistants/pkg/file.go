package pkg

import (
	"bytes"
	"encoding/json"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"io"
	"net/http"
)

type ListFilesResponse struct {
	Data   []assistants.File `json:"data"`
	Object string            `json:"object"`
}

func ListFiles(client *http.Client, baseURL string, apiKey string) (*ListFilesResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/files", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var listFilesResponse ListFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&listFilesResponse); err != nil {
		return nil, err
	}

	return &listFilesResponse, nil
}

type CreateFileRequest struct {
	File    []byte `json:"file"`
	Purpose string `json:"purpose"`
}

func CreateFile(client *http.Client, baseURL string, apiKey string, request CreateFileRequest) (*assistants.File, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/files", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var file assistants.File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	return &file, nil
}

type DeleteFileResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

func DeleteFile(client *http.Client, baseURL string, apiKey string, fileID string) (*DeleteFileResponse, error) {
	req, err := http.NewRequest("DELETE", baseURL+"/files/"+fileID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var deleteFileResponse DeleteFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteFileResponse); err != nil {
		return nil, err
	}

	return &deleteFileResponse, nil
}
