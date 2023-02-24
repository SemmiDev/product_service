package product

import (
	"context"
	"database/sql"
	"time"

	"github.com/elastic/go-elasticsearch/v8"

	"github.com/google/uuid"
)

// Product is the entity for the product table
type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Quantity    uint      `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// InitProduct is a constructor for the Product entity
func (p *Product) InitProduct() {
	now := time.Now().UTC()

	p.ID = uuid.New()
	p.CreatedAt = now
	p.UpdatedAt = now
}

// CommandHandler is the interface for all command handlers
type CommandHandler interface {
	Handle(ctx context.Context, db *sql.DB) error
}

// QueryHandler is the interface for all query handlers
type QueryHandler interface {
	Handle(ctx context.Context, es *elasticsearch.Client) (any, error)
}
