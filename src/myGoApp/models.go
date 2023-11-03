package main

type UserPreference struct {
	Username      string   `json:"username"`
	ID            int      `json:"id"`
	SortOrder     string   `json:"sortOrder"`
	HiddenDevices []string `json:"hiddenDevices"`
	Icon          []byte   
}

type Device struct {
	ID       string   `json:"device_id"`
	Name     string   `json:"display_name"`
	Position Position `json:"latest_device_point"`
	IsActive string   `json:"active_state"`
}

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type ApiResponse struct {
	Devices []Device `json:"result_list"`
}
