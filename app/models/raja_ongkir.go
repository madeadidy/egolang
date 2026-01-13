package models

type ProvinceResponse struct {
	Meta Meta       `json:"meta"`
	Data []Province `json:"data"`
}

type Meta struct {
	Message     string `json:"message"`
	Code        int    `json:"code"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type Province struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CityResponse struct {
	Meta Meta   `json:"meta"`
	Data []City `json:"data"`
}

type City struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ZipCode string `json:"zip_code"` 
}

type OngkirResponse struct {
	Meta Meta             `json:"meta"`
	Data []OngkirResult   `json:"data"`
}

type OngkirResult struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	Etd         string `json:"etd"`
}

type ShippingFeeParams struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Weight      int    `json:"weight"`
	Courier     string `json:"courier"`
}

type ShippingFeeOption struct {
	Service string `json:"service"`
	Fee     int64  `json:"fee"`
}