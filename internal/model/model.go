package model

import (
	"database/sql"
	"errors"
)

// Product is a product that we sell
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Get a product
func (p *Product) Get(db *sql.DB) error {
	return errors.New("Not implemented")
}

// Update a product
func (p *Product) Update(db *sql.DB) error {
	return errors.New("Not implemented")
}

// Delete a product
func (p *Product) Delete(db *sql.DB) error {
	return errors.New("Not implemented")
}

// Create a product
func (p *Product) Create(db *sql.DB) error {
	return errors.New("Not implemented")
}

// GetProducts gets `count` products starting at id `start`
func GetProducts(db *sql.DB, start, count int) ([]Product, error) {
	return nil, errors.New("Not implemented")
}
