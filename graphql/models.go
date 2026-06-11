package graphql

type Account struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Orders []Order `json:"orders"`
}

type Order struct {
	ID        string  `json:"id"`
	Product   string  `json:"product"`
	Price     float64 `json:"price"`
	AccountID string  `json:"accountId"`
}