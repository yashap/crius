package model

import (
	"database/sql"
)

// Product is a product that we sell
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Get a product
func (p *Product) Get(db *sql.DB) error {
	return db.QueryRow(
		`SELECT "name", "price" FROM "products" WHERE id = $1`, p.ID,
	).Scan(&p.Name, &p.Price)
}

// Update a product
func (p *Product) Update(db *sql.DB) error {
	_, err := db.Exec(
		`UPDATE "products" SET "name" = $1, "price" = $2 WHERE "id" = $3`, p.Name, p.Price, p.ID,
	)
	return err
}

// Delete a product
func (p *Product) Delete(db *sql.DB) error {
	_, err := db.Exec(
		`DELETE FROM "products" WHERE "id" = $1`, p.ID,
	)
	return err
}

// Create a product
func (p *Product) Create(db *sql.DB) error {
	err := db.QueryRow(
		`INSERT INTO "products" ("name", "price") VALUES ($1, $2) RETURNING "id"`, p.Name, p.Price,
	).Scan(&p.ID)
	return err
}

// GetProducts gets `count` products starting at id `start`
func GetProducts(db *sql.DB, start, count int) ([]Product, error) {
	rows, err := db.Query(
		`SELECT "id", "name", "price" FROM "products" LIMIT $1 OFFSET $2`,
		count,
		start,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
