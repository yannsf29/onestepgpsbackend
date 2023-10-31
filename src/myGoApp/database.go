package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func initDB(filepath string) *sql.DB {
	// Connect to the database
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create tables if they don't already exist
	createTables(db)

	return db
}

func createTables(db *sql.DB) {
	// Create the user preferences table
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS user_preferences (
            id INTEGER PRIMARY KEY,
            sort_order TEXT,
            hidden_devices TEXT,
            icon BLOB
        );
    `)

	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func getUserPreference(db *sql.DB, id int) (UserPreference, error) {
	var pref UserPreference
	var hiddenDevices string

	err := db.QueryRow("SELECT id, sort_order, hidden_devices, icon FROM user_preferences WHERE id=?", id).Scan(&pref.ID, &pref.SortOrder, &hiddenDevices, &pref.Icon)
	
	// Handle potential SQL errors.
	if err != nil {
		if err == sql.ErrNoRows {
			return pref, fmt.Errorf("No user preference found for ID %d", id)
		}
		return pref, fmt.Errorf("Database error: %v", err)
	}

	// Handle potential JSON unmarshal errors.
	if err := json.Unmarshal([]byte(hiddenDevices), &pref.HiddenDevices); err != nil {
		return pref, fmt.Errorf("Failed to unmarshal hidden devices: %v", err)
	}

    log.Printf("Fetched user preferences: %+v", pref)
	return pref, nil
}

func updateUserPreference(db *sql.DB, pref UserPreference) error {
    // Marshal the HiddenDevices slice into a JSON string
    hiddenDevicesJSON, err := json.Marshal(pref.HiddenDevices)
    if err != nil {
        log.Printf("Error marshaling hidden devices: %v", err)
        return err
    }

    // Use the UPDATE SQL statement to update the user_preferences table
    result, err := db.Exec(
        "UPDATE user_preferences SET sort_order = ?, hidden_devices = ?, icon = ? WHERE id = ?",
        pref.SortOrder, string(hiddenDevicesJSON), pref.Icon, pref.ID,
    )
    if err != nil {
        log.Printf("Error updating user preferences in database: %v", err)
        return err
    }

    // Log the marshaled hidden devices for debugging purposes
    log.Printf("Marshalled hidden devices: %s", string(hiddenDevicesJSON))
    log.Printf("Updating user preferences for ID: %d", pref.ID)

    // Check the number of rows affected by the update
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error fetching rows affected: %v", err)
        return err
    }
    if rowsAffected == 0 {
        // If no rows were affected, it means the record doesn't exist.
        log.Printf("No records were updated. The user with ID: %d might not exist in the database.", pref.ID)
        // Handle the case where the user does not exist if necessary
    } else {
        log.Printf("Updated user preferences for ID: %d", pref.ID)
    }

    return nil
}




func createUserPreference(db *sql.DB, pref UserPreference) error {
	_, err := db.Exec("INSERT INTO user_preferences(sort_order, hidden_devices, icon) VALUES (?, ?, ?)", pref.SortOrder, pref.HiddenDevices, pref.Icon)
	return err
}
