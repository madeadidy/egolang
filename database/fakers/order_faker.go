package fakers

import (
	"math/rand"
	"time"

	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrdersFaker returns a slice of sample orders to seed the database.
func OrdersFaker(db *gorm.DB) []models.Order {
    var users []models.User
    var products []models.Product
    db.Find(&users)
    db.Find(&products)

    rand.Seed(time.Now().UnixNano())

    var out []models.Order
    if len(users) == 0 || len(products) == 0 {
        return out
    }

    // create 10 sample orders
    for i := 0; i < 10; i++ {
        u := users[rand.Intn(len(users))]
        // pick 1..3 items
        itemCount := 1 + rand.Intn(3)
        var items []models.OrderItem
        grand := decimal.NewFromInt(0)
        for j := 0; j < itemCount; j++ {
            p := products[rand.Intn(len(products))]
            qty := 1 + rand.Intn(5)
            base := p.Price
            qtyDec := decimal.NewFromInt(int64(qty))
            sub := base.Mul(qtyDec)
            oi := models.OrderItem{
                ID:         uuid.New().String(),
                ProductID:  p.ID,
                Qty:        qty,
                BasePrice:  base,
                BaseTotal:  sub,
                SubTotal:   sub,
                Sku:        p.Sku,
                Name:       p.Name,
                Weight:     p.Weight,
                CreatedAt:  time.Time{},
                UpdatedAt:  time.Time{},
            }
            items = append(items, oi)
            grand = grand.Add(sub)
        }

        // small random shipping cost
        ship := decimal.NewFromFloat(float64(rand.Intn(20000)))
        grand = grand.Add(ship)

        ord := models.Order{
            ID:         uuid.New().String(),
            UserID:     u.ID,
            OrderItems: items,
            Status:     0,
            OrderDate:  time.Now().AddDate(0, 0, -rand.Intn(30)),
            BaseTotalPrice: grand,
            GrandTotal:  grand,
            PaymentStatus: "PAID",
            CreatedAt:   time.Time{},
        }
        out = append(out, ord)
    }

    return out
}
