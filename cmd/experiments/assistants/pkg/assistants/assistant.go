package assistants

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
)

func CreateAssistant(apiKey string, assistantData Assistant) (*Assistant, error) {
	url := "https://api.openai.com/v1/assistants"
	jsonBody, err := json.Marshal(assistantData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, err
	}

	return &assistant, nil
}

func RetrieveAssistant(apiKey, assistantID string) (*Assistant, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/assistants/%s", assistantID)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, err
	}

	return &assistant, nil
}

func ModifyAssistant(apiKey, assistantID string, updateData Assistant) (*Assistant, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/assistants/%s", assistantID)
	jsonBody, err := json.Marshal(updateData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, err
	}

	return &assistant, nil
}

func DeleteAssistant(apiKey, assistantID string) error {
	url := fmt.Sprintf("https://api.openai.com/v1/assistants/%s", assistantID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to delete assistant")
	}

	return nil
}

type PaginationResponse struct {
	Object  string      `json:"object"`
	Data    []Assistant `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`
}

func ListAssistants(apiKey, after string, limit int) ([]Assistant, bool, error) {
	url := "https://api.openai.com/v1/assistants"
	if limit > 0 || after != "" {
		url += "?"
		queryParts := []string{}
		if limit > 0 {
			queryParts = append(queryParts, fmt.Sprintf("limit=%d", limit))
		}
		if after != "" {
			queryParts = append(queryParts, "after="+after)
		}
		url += strings.Join(queryParts, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var page PaginationResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, false, err
	}

	return page.Data, page.HasMore, nil
}
