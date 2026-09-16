package dto

type Report struct {
	Violations   []Violation `json:"violations"`
	Count        int         `json:"count"`
	WarningCount int         `json:"warning_count"`
}
