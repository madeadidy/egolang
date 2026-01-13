package models

type CustomProductImage struct {
	ID              int
	CustomProductID int
	ImageURL        string
	IsMain          bool
}
