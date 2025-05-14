package models

type User struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	PasswordHash      string `json:"-"`
	Role              string `json:"role"` // user, admin, tech
	SubscriptionPlan  string `json:"subscription_plan"`
	SubscriptionStart string `json:"subscription_start"`
	SubscriptionEnd   string `json:"subscription_end"`
}