package product

import (
	"context"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type GetProductByIdQuery struct {
	ID uuid.UUID
}

func (q *GetProductByIdQuery) Handle(ctx context.Context, es *elasticsearch.Client) (any, error) {
	getReq := esapi.GetRequest{
		Index:      "products",
		DocumentID: q.ID.String(),
	}

	res, err := getReq.Do(context.Background(), es)
	if err != nil {
		log.Printf("failed do get request: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrorProductNotFound
	}

	var rawData map[string]any
	err = json.NewDecoder(res.Body).Decode(&rawData)
	if err != nil {
		return nil, err
	}

	source := rawData["_source"].(map[string]any)
	productResult := EsSourceToProduct(source)

	return &productResult, nil
}
