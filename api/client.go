package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"2023-go/scoring"
	"2023-go/types"
)

type Client struct {
	apiKey  string
	baseUrl string
	client  http.Client
}

type SubmitSolutionRequest struct {
	Locations map[string]map[string]int `json:"locations"`
}

func NewClient(apiKey string, baseUrl string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseUrl: baseUrl,
		client:  http.Client{},
	}
}

func (c *Client) GetMapData(mapName string) (*types.MapData, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/Game/getMapData?mapName=%s", c.baseUrl, mapName), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("x-api-key", c.apiKey)
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var mapData types.MapData
	err = json.NewDecoder(resp.Body).Decode(&mapData)
	if err != nil {
		return nil, err
	}

	return &mapData, nil
}

func (c *Client) GetGeneralData() (*types.GeneralGameData, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/Game/getGeneralGameData", c.baseUrl), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("x-api-key", c.apiKey)
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var generalData types.GeneralGameData
	err = json.NewDecoder(resp.Body).Decode(&generalData)
	if err != nil {
		return nil, err
	}

	return &generalData, nil
}

func (c *Client) SubmitSolution(mapName string, solution scoring.ScoredSolution) (*types.GameData, error) {
	submitLocations := make(map[string]map[string]int)
	for _, loc := range solution.Locations {
		submitLocations[loc.Location.Name] = map[string]int{
			"freestyle3100Count": loc.F3,
			"freestyle9100Count": loc.F9,
		}
	}

	submitSolutionRequest := SubmitSolutionRequest{
		Locations: submitLocations,
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(submitSolutionRequest)
	if err != nil {
		return nil, err
	}

	file, err := os.Create("solution.json")
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(file).Encode(submitSolutionRequest); err != nil {
		return nil, err
	}

	file.Close()

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/Game/submitSolution?mapName=%s", c.baseUrl, mapName), body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("x-api-key", c.apiKey)
	request.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var gameData types.GameData
	err = json.NewDecoder(resp.Body).Decode(&gameData)
	if err != nil {
		return nil, err
	}

	return &gameData, nil
}
