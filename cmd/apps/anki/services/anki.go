package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

const ankiConnectURL = "http://localhost:8765"

type AnkiService struct {
	client *http.Client
	logger zerolog.Logger
}

type AnkiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

type AnkiResponse struct {
	Result interface{} `json:"result"`
	Error  *string     `json:"error"`
}

func NewAnkiService(logger zerolog.Logger) *AnkiService {
	return &AnkiService{
		client: &http.Client{},
		logger: logger,
	}
}

func (s *AnkiService) GetDecks() ([]string, error) {
	req := AnkiRequest{
		Action:  "deckNames",
		Version: 6,
	}

	var resp AnkiResponse
	if err := s.doRequest(req, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf(*resp.Error)
	}

	decks, ok := resp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	deckNames := make([]string, len(decks))
	for i, deck := range decks {
		deckNames[i] = deck.(string)
	}

	return deckNames, nil
}

func (s *AnkiService) GetCardsInDeck(deckName string) ([]map[string]interface{}, error) {
	// First find cards in deck
	findReq := AnkiRequest{
		Action:  "findCards",
		Version: 6,
		Params: map[string]string{
			"query": fmt.Sprintf("\"deck:%s\"", deckName),
		},
	}

	var findResp AnkiResponse
	if err := s.doRequest(findReq, &findResp); err != nil {
		return nil, err
	}

	if findResp.Error != nil {
		return nil, fmt.Errorf(*findResp.Error)
	}

	cardIDs, ok := findResp.Result.([]interface{})
	if !ok || len(cardIDs) == 0 {
		return nil, nil
	}

	// Then get card info
	infoReq := AnkiRequest{
		Action:  "cardsInfo",
		Version: 6,
		Params: map[string]interface{}{
			"cards": cardIDs,
		},
	}

	var infoResp AnkiResponse
	if err := s.doRequest(infoReq, &infoResp); err != nil {
		return nil, err
	}

	if infoResp.Error != nil {
		return nil, fmt.Errorf(*infoResp.Error)
	}

	cards, ok := infoResp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	result := make([]map[string]interface{}, len(cards))
	for i, card := range cards {
		cardMap, ok := card.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected card format")
		}
		result[i] = cardMap
	}

	return result, nil
}

func (s *AnkiService) GetModels() ([]map[string]interface{}, error) {
	// First get model names and IDs
	namesReq := AnkiRequest{
		Action:  "modelNamesAndIds",
		Version: 6,
	}

	var namesResp AnkiResponse
	if err := s.doRequest(namesReq, &namesResp); err != nil {
		return nil, err
	}

	if namesResp.Error != nil {
		return nil, fmt.Errorf(*namesResp.Error)
	}

	modelsMap, ok := namesResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format for modelNamesAndIds")
	}

	// Get detailed info for each model
	var modelNames []string
	for name := range modelsMap {
		modelNames = append(modelNames, name)
	}

	findReq := AnkiRequest{
		Action:  "findModelsByName",
		Version: 6,
		Params: map[string]interface{}{
			"modelNames": modelNames,
		},
	}

	var findResp AnkiResponse
	if err := s.doRequest(findReq, &findResp); err != nil {
		return nil, err
	}

	if findResp.Error != nil {
		return nil, fmt.Errorf(*findResp.Error)
	}

	models, ok := findResp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format for findModelsByName")
	}

	result := make([]map[string]interface{}, len(models))
	for i, model := range models {
		modelMap, ok := model.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected model format")
		}
		result[i] = modelMap
	}

	return result, nil
}

func (s *AnkiService) doRequest(req AnkiRequest, resp *AnkiResponse) error {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Log the full request in debug mode
	if s.logger.Debug().Enabled() {
		prettyReq, err := json.MarshalIndent(req, "", "  ")
		if err == nil {
			s.logger.Debug().
				Str("action", req.Action).
				RawJSON("request", prettyReq).
				Msg("Sending request to Anki")
		}
	}

	httpReq, err := http.NewRequest("POST", ankiConnectURL, bytes.NewBuffer(jsonReq))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return err
	}

	// Log the full response in debug mode
	if s.logger.Debug().Enabled() {
		prettyResp, err := json.MarshalIndent(resp, "", "  ")
		if err == nil {
			s.logger.Debug().
				Str("action", req.Action).
				RawJSON("response", prettyResp).
				Msg("Received response from Anki")
		}
	}

	return nil
}
