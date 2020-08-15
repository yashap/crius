package model

// Service represents a service
type Service struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" binding:"required"`
}
