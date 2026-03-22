package models

// PaginatedResponse represents a paginated list of files
type PaginatedResponse struct {
	Files       []FileInfo `json:"files"`
	Total       int        `json:"total"`
	Page        int        `json:"page"`
	Limit       int        `json:"limit"`
	HasMore     bool       `json:"hasMore"`
	TotalPages  int        `json:"totalPages"`
}
