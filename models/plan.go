package models

type Plan struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	RemoteCalls int     `json:"remote_calls"`
	OnsiteCalls int     `json:"onsite_calls"`
}