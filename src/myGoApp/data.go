package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
