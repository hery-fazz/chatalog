package merchant

import "context"

type Service interface {
	GenerateBrochure(ctx context.Context, merchantPhone string, productNames []string) (string, error)
}

type Repository interface {
	GetMerchantByPhone(ctx context.Context, phone string) (*Merchant, error)
	GetProductsByMerchantID(ctx context.Context, merchantID string) ([]Product, error)
}
