package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
    "strings"
    "encoding/base64"
    "strconv"
)

var (
    apiKey string
    db     *sql.DB
)

func init() {
	apiKey = os.Getenv("ONESTEPGPS_API_KEY")
	if apiKey == "" {
		panic("ONESTEPGPS_API_KEY environment variable not set!")
	}
}

func main() {
    db = initDB("./preferences.db")
    defer db.Close()

    http.HandleFunc("/", handler)
    http.HandleFunc("/preferences", handleGetUserPreference) // GET request
    http.HandleFunc("/preferences/update", handleUpdateUserPreference) // POST request

    log.Fatal(http.ListenAndServe(":8081", nil))
}

func contains(slice []string, val string) bool {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}

func handler(w http.ResponseWriter, r *http.Request) {
    setCORSHeaders(w) 

    data, err := FetchData(apiKey)
    if err != nil {
        log.Printf("Error fetching data: %v", err)
        http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
        return
    }
    
    pref, err := getUserPreference(db, 1)
    if err != nil {
        log.Printf("Error fetching user preferences: %v", err)
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
        log.Printf("Error converting data to JSON: %v", err)
        http.Error(w, "Failed to convert data to JSON", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}


func handleGetUserPreference(w http.ResponseWriter, r *http.Request) {
    setCORSHeaders(w)

    // Extract user ID from the URL path.
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) != 3 {
        http.Error(w, "Invalid URL format", http.StatusBadRequest)
        return
    }
    idStr := parts[2]

    // Convert the idStr to an integer
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid user ID format", http.StatusBadRequest)
        return
    }

    // Define pref here before you attempt to use it.
    pref, err := getUserPreference(db, id)
    if err != nil {
        http.Error(w, "Failed to fetch user preferences", http.StatusInternalServerError)
        return
    }

    // Convert the icon byte slice to a Base64 string before you marshal it into JSON.
    base64Icon := base64.StdEncoding.EncodeToString(pref.Icon)
    pref.Icon = []byte(base64Icon)

    response, err := json.Marshal(pref)
    if err != nil {
        log.Printf("Failed to marshal user preferences: %v", err)
        http.Error(w, "Failed to convert user preferences to JSON", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}

func handleUpdateUserPreference(w http.ResponseWriter, r *http.Request) {
    setCORSHeaders(w)

    // Respond to the preflight OPTIONS request
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // Handle only POST requests for updating preferences
    if r.Method != "POST" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusMethodNotAllowed)
        w.Write([]byte(`{"error": "Method not allowed"}`))
        return
    }

    // Define pref inside this function's scope.
    var pref UserPreference

    // Decode the request body directly without reading it to a separate variable first.
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&pref) // Use 'err' variable that is already declared
    if err != nil {
        log.Printf("Error decoding user preferences: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(`{"error": "Bad request data"}`))
        return
    }

    // Log the decoded user preference for debugging
    log.Printf("Decoded user preferences: %+v", pref)

    // Update the user preference in the database
    err = updateUserPreference(db, pref)
    if err != nil {
        log.Printf("Error updating user preferences: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error": "Failed to update user preferences"}`))
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Preference updated successfully"}`))
}


func setCORSHeaders(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept")
}


