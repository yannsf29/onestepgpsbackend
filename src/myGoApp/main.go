package main

import (
	"database/sql"
	"net/http"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	apiKey := os.Getenv("ONESTEPGPS_API_KEY")
	if apiKey == "" {
		panic("ONESTEPGPS_API_KEY environment variable not set!")
	}

	db, err := sql.Open("sqlite3", "./preferences.db")
	if err != nil {
		panic("Failed to open database: " + err.Error())
	}
	defer db.Close()

	// Initialize the handlers with dependencies
	deps := &HandlerDependencies{
		DB:     db,
		ApiKey: apiKey,
	}

	// Setup routes with the new handler methods
	http.HandleFunc("/", deps.Handler)
	http.HandleFunc("/preferences", deps.HandleGetUserPreference) // GET request
	http.HandleFunc("/preferences/update", deps.HandleUpdateUserPreference) // POST request

	panic(http.ListenAndServe(":8081", nil))
}
