package product

type Product struct {
	UserID   string  `json:"user_id,omitempty"`
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name"`
	Price    float64 `json:"price,omitempty"`
	Currency string  `json:"currency,omitempty"`
	ImageURL *string `json:"image_url,omitempty"`
}

type ListOutput struct {
	Items      []Product `json:"items"`
	Total      int       `json:"total"`
	NextOffset *int      `json:"next_offset,omitempty"`
}
