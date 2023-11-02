package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func createTables(db *sql.DB) {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS user_preferences (
            id INTEGER PRIMARY KEY,
            sort_order TEXT,
            hidden_devices TEXT,
            icon BLOB
        );
    `)
	if err != nil {
		panic(fmt.Sprintf("Failed to create tables: %v", err))
	}
}

func getUserPreference(db *sql.DB, id int) (UserPreference, error) {
	var pref UserPreference
	var hiddenDevices string

	err := db.QueryRow("SELECT id, sort_order, hidden_devices, icon FROM user_preferences WHERE id=?", id).Scan(&pref.ID, &pref.SortOrder, &hiddenDevices, &pref.Icon)

	if err != nil {
		if err == sql.ErrNoRows {
			return pref, fmt.Errorf("No user preference found for ID %d", id)
		}
		return pref, fmt.Errorf("Database error: %v", err)
	}

	if err := json.Unmarshal([]byte(hiddenDevices), &pref.HiddenDevices); err != nil {
		return pref, fmt.Errorf("Failed to unmarshal hidden devices: %v", err)
	}

	return pref, nil
}

func updateUserPreference(db *sql.DB, pref UserPreference) error {
	hiddenDevicesJSON, err := json.Marshal(pref.HiddenDevices)
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"UPDATE user_preferences SET sort_order = ?, hidden_devices = ?, icon = ? WHERE id = ?",
		pref.SortOrder, string(hiddenDevicesJSON), pref.Icon, pref.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Handle the case where no rows were affected
	}

	return nil
}

func createUserPreference(db *sql.DB, pref UserPreference) error {
	hiddenDevicesJSON, err := json.Marshal(pref.HiddenDevices)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_preferences(sort_order, hidden_devices, icon) VALUES (?, ?, ?)",
		pref.SortOrder, string(hiddenDevicesJSON), pref.Icon)
	return err
}
