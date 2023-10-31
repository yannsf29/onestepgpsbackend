package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type Device struct {
	ID       string   `json:"device_id"`
	Name     string   `json:"display_name"`
	Position Position `json:"latest_device_point"`
	IsActive string    `json:"active_state"`
}

type ApiResponse struct {
	Devices []Device `json:"result_list"`
}

func FetchData(apiKey string) (ApiResponse, error) {
	resp, err := http.Get("https://track.onestepgps.com/v3/api/public/device?latest_point=true&api-key=" + apiKey)
	if err != nil {
		return ApiResponse{}, fmt.Errorf("error making http request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ApiResponse{}, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ApiResponse{}, fmt.Errorf("error reading response body: %v", err)
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return ApiResponse{}, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return apiResponse, nil
}
