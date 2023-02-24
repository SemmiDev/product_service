package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	nats_messaging "github.com/semmidev/product_service/nats"
	"github.com/semmidev/product_service/product"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	postgres, err := sql.Open("postgres", "postgres://root:secret@localhost:5432/products?sslmode=disable")
	if err != nil {
		log.Printf("failed to connect to write database: %v", err)
	}

	if err := postgres.Ping(); err != nil {
		log.Printf("failed to ping write database: %v", err)
	}

	log.Println("👉 connected to postgres database")

	// https://www.elastic.co/guide/en/elasticsearch/client/go-api/master/connecting.html#connecting-without-security
	esCfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
		Username: "sammi",
		Password: "sammi",
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Fatalf("error creating the client: %v", err)
	}

	// Buat request untuk membuat index
	productsIndex := "products"
	createIndexRequest := map[string]any{
		"settings": map[string]any{
			"number_of_shards": 1,
			"number_of_replicas": 0,
		},
	}

	var bytesBuffer bytes.Buffer
	if err := json.NewEncoder(&bytesBuffer).Encode(createIndexRequest); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	res, err := es.Indices.Create(productsIndex, es.Indices.Create.WithBody(&bytesBuffer))
	if err != nil {
		fmt.Printf("Error creating the index: %s", err)
		return
	}
	defer res.Body.Close()

	res, err = es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response: %v", res.String())
	}

	if res.StatusCode == http.StatusOK {
		log.Printf("👉 Index %s is created", productsIndex)
	}

	log.Println("👉 elastic search Connection succeeded")

	natsMessaging, err := nats_messaging.NewNatsMessaging()
	if err != nil {
		log.Fatalf("failed to connect to NATS messaging: %v", err)
	}
	defer natsMessaging.Close()

	natsMessaging.Subscribe(nats_messaging.ProductCreatedEvent, es)
	service := product.NewService(postgres, es, natsMessaging)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/products", func(w http.ResponseWriter, r *http.Request) {
		var payload product.Product
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			log.Printf("Failed to decode product data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = service.CreateProduct(r.Context(), &payload)
		if err != nil {
			log.Printf("Failed to create product: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	})

	r.Get("/products", func(w http.ResponseWriter, r *http.Request) {
		products, err := service.GetAllProducts(r.Context())
		if err != nil {
			log.Printf("Failed to get all products: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})

	r.Get("/products/search", func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("term")
		products, err := service.SearchByTerm(r.Context(), term)
		if err != nil {
			log.Printf("Failed to search products: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})

	r.Get("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		productId := chi.URLParam(r, "id")

		response, err := service.GetProductById(r.Context(), productId)
		if err != nil {
			if errors.Is(err, product.ErrInvalidId) {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else if errors.Is(err, product.ErrorProductNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	runDBMigration("file://migrations", "postgres://root:secret@localhost:5432/products?sslmode=disable")
	log.Println("👉 migration postgres db success")

	log.Println("👉 starting server on port 3030")
	log.Fatal(http.ListenAndServe(":3030", r))
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatalf("cannot create new migrate instance: %v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("cannot run migration up: %v", err)
	}
}
