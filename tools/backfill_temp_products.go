package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/codeuiprogramming/e-commerce/database"
)

func main() {
    // load .env if present
    _ = godotenv.Load()

    database.Init()

    // Backfill products with slug starting with custom- to be temporary
    res := database.DB.Exec("UPDATE products SET is_temporary = true WHERE slug LIKE 'custom-%' AND (is_temporary IS NULL OR is_temporary = false);")
    if res.Error != nil {
        log.Fatalf("backfill by slug failed: %v", res.Error)
    }
    fmt.Printf("Updated by slug rows affected: %d\n", res.RowsAffected)

    // Also mark products that have images under uploads/custom/
    res2 := database.DB.Exec("UPDATE products SET is_temporary = true WHERE id IN (SELECT product_id FROM product_images WHERE path LIKE '%uploads/custom/%') AND (is_temporary IS NULL OR is_temporary = false);")
    if res2.Error != nil {
        log.Fatalf("backfill by image path failed: %v", res2.Error)
    }
    fmt.Printf("Updated by image-path rows affected: %d\n", res2.RowsAffected)

    fmt.Println("Backfill complete.")
}
