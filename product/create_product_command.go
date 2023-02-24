package product

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type CreateProductCommand struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Quantity    uint      `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *CreateProductCommand) Handle(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO products (id, name, description, category, price, quantity, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.Name, c.Description, c.Category, c.Price,  c.Quantity, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		log.Printf("failed to insert product to read database: %v", err)
		return err
	}

	log.Printf("new product with id %v created", c.ID)
	return nil
}
