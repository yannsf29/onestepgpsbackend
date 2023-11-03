package main

import (
	"database/sql" 
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
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

	
	userID, err := getUserIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid user ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	
	var pref UserPreference
	err = json.NewDecoder(r.Body).Decode(&pref)
	if err != nil {
		http.Error(w, "Bad request data: "+err.Error(), http.StatusBadRequest)
		return
	}

	
	pref.ID = userID


	err = updateUserPreference(deps.DB, pref)
	if err != nil {
		http.Error(w, "Failed to update user preferences: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Preference updated successfully"}`))
}

func (deps *HandlerDependencies) HandleGetUserPreferenceByUsername(w http.ResponseWriter, r *http.Request) {
    setCORSHeaders(w)

   
    username, err := getUsernameFromURL(r.URL.Path)
    if err != nil {
        http.Error(w, "Invalid username: "+err.Error(), http.StatusBadRequest)
        return
    }

    
    pref, err := getUserPreferenceByUsername(deps.DB, username)
    if err != nil {
        
        if err == sql.ErrNoRows {
            http.Error(w, "Preferences not found for the given username", http.StatusNotFound)
        } else {
            http.Error(w, "Server error: "+err.Error(), http.StatusInternalServerError)
        }
        return
    }

    
    base64Icon := base64.StdEncoding.EncodeToString(pref.Icon)
    pref.Icon = []byte(base64Icon) 

   
    response, err := json.Marshal(pref)
    if err != nil {
        http.Error(w, "Failed to convert user preferences to JSON: "+err.Error(), http.StatusInternalServerError)
        return
    }

  
    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}

func getUsernameFromURL(path string) (string, error) {
    parts := strings.Split(strings.TrimPrefix(path, "/preferences/by-username/"), "/")
    if len(parts) != 1 {
        return "", fmt.Errorf("Invalid URL format")
    }
    username, err := url.QueryUnescape(parts[0])
    if err != nil {
        return "", fmt.Errorf("Invalid username format: %v", err)
    }
    return username, nil
}

func getUserIDFromURL(path string) (int, error) {
	
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("Invalid URL format")
	}

	
	idStr := parts[len(parts)-1]

	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid user ID format: %v", err)
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
