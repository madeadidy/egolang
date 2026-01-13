package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/codeuiprogramming/e-commerce/database"
)

func main() {
    _ = godotenv.Load()
    database.Init()

    type Row struct {
        CartItemID string
        ProductID  string
        Path       string
    }

    var rows []Row
    err := database.DB.Raw(`SELECT ci.id as cart_item_id, ci.product_id, pi.path
        FROM cart_items ci
        LEFT JOIN product_images pi ON pi.product_id = ci.product_id
        ORDER BY ci.created_at DESC
        LIMIT 100`).Scan(&rows).Error
    if err != nil {
        log.Fatalf("query failed: %v", err)
    }

    if len(rows) == 0 {
        fmt.Println("No cart items found in DB")
        return
    }

    for _, r := range rows {
        fmt.Printf("cart_item=%s product=%s path=%s\n", r.CartItemID, r.ProductID, r.Path)
    }
}
