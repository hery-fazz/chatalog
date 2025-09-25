package merchant

type Merchant struct {
	ID    string
	Name  string
	Phone string
}

type Product struct {
	ID         string
	MerchantID string
	Name       string
	Price      float64
}
