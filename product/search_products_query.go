package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchProductQuery struct {
	Term string
}

func (q *SearchProductQuery) Handle(ctx context.Context, esClient *elasticsearch.Client) (any, error) {
	pageSize := 25
	pageNum := 1

	// query := map[string]any{
	// 	"query": map[string]any{
	// 		"bool": map[string]any{
	// 			"must": []any{
	// 				map[string]any{
	// 					"match": map[string]any{
	// 						"name": q.Term,
	// 					},
	// 				},
	// 				map[string]any{
	// 					"match": map[string]any{
	// 						"price": 12000000,
	// 					},
	// 				},
	// 			},
	// 		},

	// 	},
	// 	"from": (pageNum - 1) * pageSize,
	// 	"size": pageSize,
	// }

	query := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  q.Term,
				"type":   "best_fields",
				"fields": []string{"name", "description"},
			},
		},
		"from": (pageNum - 1) * pageSize,
		"size": pageSize,
	}

	// query := map[string]any{
	// 	"query": map[string]any{
	// 		"match": map[string]any{
	// 			"name": q.Term,
	// 		},
	// 	},
	// 	"from": (pageNum - 1) * pageSize,
	// 	"size": pageSize,
	// }

	var buf []byte
	buf, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("products"),
		esClient.Search.WithBody(bytes.NewReader(buf)),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		return []*Product{}, err
	}
	defer res.Body.Close()

	// Membaca hasil response search
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return []*Product{}, err
	}
	fmt.Println(result)

	data := result["hits"].(map[string]any)["hits"].([]any)
	if len(data) == 0 {
		return []*Product{}, nil
	}

	var products []*Product
	for _, item := range data {
		raw := item.(map[string]any)["_source"].(map[string]any)
		product := EsSourceToProduct(raw)
		products = append(products, &product)
	}

	return products, nil
}
