package merchant

import (
	"context"

	"github.com/defryfazz/fazztalog/internal/ai"
)

type service struct {
	repo     Repository
	aiEngine ai.Engine
}

func NewService(repo Repository, aiEngine ai.Engine) Service {
	return &service{
		repo:     repo,
		aiEngine: aiEngine,
	}
}

func (s *service) GenerateBrochure(ctx context.Context, merchantPhone string) (string, error) {
	merchant, err := s.repo.GetMerchantByPhone(ctx, merchantPhone)
	if err != nil {
		return "", err
	}

	products, err := s.repo.GetProductsByMerchantID(ctx, merchant.ID)
	if err != nil {
		return "", err
	}

	aiProducts := make([]ai.Product, 0, len(products))
	for _, p := range products {
		aiProducts = append(aiProducts, ai.Product{
			Name:  p.Name,
			Price: p.Price,
		})
	}
	brochureDetails := ai.BrochureDetails{
		MerchantName: merchant.Name,
		Products:     aiProducts,
	}

	return s.aiEngine.GenerateBrochure(ctx, brochureDetails)
}
