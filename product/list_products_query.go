package product

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ListProductsQuery struct{}

func (q *ListProductsQuery) Handle(ctx context.Context, es *elasticsearch.Client) (any, error) {
	req := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
	}

	res, err := req.Do(ctx, es)
	if err != nil {
		log.Printf("failed do get request: %v", err)
		return []*Product{}, err
	}
	defer res.Body.Close()

	var rawData map[string]any
	if err := json.NewDecoder(res.Body).Decode(&rawData); err != nil {
		log.Printf("error parsing the response body: %s", err)
		return []*Product{}, err
	}

	// slice of map
	hits := rawData["hits"].(map[string]any)["hits"].([]any)

	if len(hits) == 0 {
		return []*Product{}, nil
	}

	// we get the source of each hit
	source := make([]map[string]any, len(hits))
	for i, hit := range hits {
		source[i] = hit.(map[string]any)["_source"].(map[string]any)
	}

	// we transform the source into a slice of Product
	products := make([]*Product, len(source))
	for i, s := range source {
		product := EsSourceToProduct(s)
		products[i] = &product
	}

	return products, nil
}
