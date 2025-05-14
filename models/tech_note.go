package models

type TechNote struct {
	ID           int    `json:"id"`
	RequestID    int    `json:"request_id"`
	TechnicianID int    `json:"technician_id"`
	Note         string `json:"note"`
	CreatedAt    string `json:"created_at"`
}