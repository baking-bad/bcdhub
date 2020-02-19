package models

// Project - entity of project
type Project struct {
	ID    string `json:"-"`
	Alias string `json:"alias,omitempty"`
}
