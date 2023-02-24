package product

import (
	"time"

	"github.com/google/uuid"
)

func EsSourceToProduct(source map[string]any) Product {
	p := Product{}

	id := source["id"].(string)
	createdAt := source["created_at"].(string)
	updateAt := source["updated_at"].(string)
	quantity := source["quantity"].(float64)

	p.ID = uuid.MustParse(id)
	p.Name = source["name"].(string)
	p.Description = source["description"].(string)
	p.Category = source["category"].(string)
	p.Price = source["price"].(float64)
	p.Quantity = uint(quantity)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updateAt)

	return p
}
