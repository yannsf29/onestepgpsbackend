package main

type UserPreference struct {
	ID           int      `json:"id"`
    SortOrder    string   `json:"sortOrder"`
    HiddenDevices []string `json:"HiddenDevices"`
	Icon         []byte
}
