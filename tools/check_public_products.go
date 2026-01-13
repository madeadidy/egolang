package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/codeuiprogramming/e-commerce/database"
)

type Product struct {
    ID   uint
    Slug string
    Name string
}

func main() {
    _ = godotenv.Load()
    database.Init()

    var products []Product
    // find products with slug like custom-% that are visible (not temporary)
    err := database.DB.Raw("SELECT id, slug, name FROM products WHERE slug LIKE 'custom-%' AND (is_temporary IS NULL OR is_temporary = false) ORDER BY id DESC LIMIT 50").Scan(&products).Error
    if err != nil {
        log.Fatalf("query failed: %v", err)
    }

    if len(products) == 0 {
        fmt.Println("No custom products visible in public listing.")
        return
    }

    fmt.Printf("Found %d custom product(s) visible:\n", len(products))
    for _, p := range products {
        fmt.Printf("- id=%d slug=%s name=%s\n", p.ID, p.Slug, p.Name)
    }
}
