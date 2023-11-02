package main

import (
	"database/sql" 
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"fmt"
)

type HandlerDependencies struct {
	DB     *sql.DB
	ApiKey string
}

func (deps *HandlerDependencies) Handler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	data, err := FetchData(deps.ApiKey)
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Get the user ID from the URL query, default to 1 if not present
	userID := 1
	if idParam, ok := r.URL.Query()["id"]; ok {
		if len(idParam) > 0 {
			userID, err = strconv.Atoi(idParam[0])
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}
		}
	}

	pref, err := getUserPreference(deps.DB, userID)
	if err != nil {
		http.Error(w, "Failed to fetch user preferences", http.StatusInternalServerError)
		return
	}

	var filteredDevices []Device
	for _, device := range data.Devices {
		if !contains(pref.HiddenDevices, device.ID) {
			filteredDevices = append(filteredDevices, device)
		}
	}

	response, err := json.Marshal(ApiResponse{Devices: filteredDevices})
	if err != nil {
		http.Error(w, "Failed to convert data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (deps *HandlerDependencies) HandleGetUserPreference(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	userID, err := getUserIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pref, err := getUserPreference(deps.DB, userID)
	if err != nil {
		http.Error(w, "Failed to fetch user preferences", http.StatusInternalServerError)
		return
	}

	base64Icon := base64.StdEncoding.EncodeToString(pref.Icon)
	pref.Icon = []byte(base64Icon)

	response, err := json.Marshal(pref)
	if err != nil {
		http.Error(w, "Failed to convert user preferences to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (deps *HandlerDependencies) HandleUpdateUserPreference(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var pref UserPreference
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pref)
	if err != nil {
		http.Error(w, "Bad request data", http.StatusBadRequest)
		return
	}

	err = updateUserPreference(deps.DB, pref)
	if err != nil {
		http.Error(w, "Failed to update user preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Preference updated successfully"}`))
}

// Helper function to extract userID from the URL path
func getUserIDFromURL(path string) (int, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		return 0, fmt.Errorf("Invalid URL format")
	}
	idStr := parts[2]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid user ID format")
	}

	return id, nil
}

func contains(slice []string, val string) bool {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, X-Requested-With")
}
