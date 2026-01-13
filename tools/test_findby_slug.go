package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/codeuiprogramming/e-commerce/database"
)

func main() {
    _ = godotenv.Load()
    database.Init()

    slug := os.Args[1]
    pm := models.Product{}
    p, err := pm.FindBySlug(database.DB, slug)
    if err != nil {
        log.Fatalf("FindBySlug error: %v", err)
    }
    fmt.Printf("Found product: %s (ID=%s)\n", p.Name, p.ID)
}
