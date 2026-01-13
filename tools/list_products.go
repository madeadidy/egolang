package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/codeuiprogramming/e-commerce/database"
)

type P struct {
    ID   string
    Slug string
    Name string
}

func main() {
    _ = godotenv.Load()
    database.Init()

    var rows []P
    err := database.DB.Raw("SELECT id, slug, name FROM products WHERE is_temporary = false ORDER BY created_at DESC LIMIT 20").Scan(&rows).Error
    if err != nil {
        log.Fatalf("query failed: %v", err)
    }

    for _, r := range rows {
        fmt.Printf("id=%s slug=%s name=%s\n", r.ID, r.Slug, r.Name)
    }
}
