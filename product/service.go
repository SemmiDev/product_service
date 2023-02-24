package product

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	nats_messaging "github.com/semmidev/product_service/nats"

	"github.com/google/uuid"
)

type service struct {
	postgresDB    *sql.DB
	elasticClient *elasticsearch.Client
	natsMessaging *nats_messaging.NatsMessaging
}

func NewService(postgresDB *sql.DB, elasticClient *elasticsearch.Client, natsMessaging *nats_messaging.NatsMessaging) *service {
	return &service{postgresDB: postgresDB, elasticClient: elasticClient, natsMessaging: natsMessaging}
}

func (s *service) CreateProduct(ctx context.Context, p *Product) error {
	p.InitProduct()

	cmd := &CreateProductCommand{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Quantity:    p.Quantity,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	err := cmd.Handle(ctx, s.postgresDB)
	if err != nil {
		log.Printf("failed to create new product: %v", err)
		return err
	}

	log.Printf("new product created: %v", cmd.ID)

	err = s.natsMessaging.Publish(nats_messaging.ProductCreatedEvent, nats_messaging.ProductPayload{
		ID:        cmd.ID,
		Name:      cmd.Name,
		Description: cmd.Description,
		Category: cmd.Category,
		Price:     cmd.Price,
		Quantity:  cmd.Quantity,
		CreatedAt: cmd.CreatedAt,
		UpdatedAt: cmd.UpdatedAt,
	})
	if err != nil {
		log.Printf("failed to publish product created event: %v\n", err)
		return err
	}

	log.Printf("product created event published: %v", cmd.ID)
	return nil
}

func (s *service) SearchByTerm(ctx context.Context, term string) ([]*Product, error) {
	qry := &SearchProductQuery{Term: term}
	result, err := qry.Handle(ctx, s.elasticClient)
	if err != nil {
		return nil, err
	}

	var products []*Product
	products = append(products, result.([]*Product)...)
	return products, nil
}

func (s *service) GetAllProducts(ctx context.Context) ([]*Product, error) {
	qry := &ListProductsQuery{}
	result, err := qry.Handle(ctx, s.elasticClient)
	if err != nil {
		return nil, err
	}

	var products []*Product
	products = append(products, result.([]*Product)...)
	return products, nil
}

var (
	ErrInvalidId = errors.New("invalid id")
	ErrorProductNotFound = errors.New("product not found")
)

func (s *service) GetProductById(ctx context.Context, id string) (*Product, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidId
	}

	qry := &GetProductByIdQuery{ID: parsedId}
	result, err := qry.Handle(ctx, s.elasticClient)
	if result == nil || err != nil {
		return nil, err
	}

	return result.(*Product), nil
}
