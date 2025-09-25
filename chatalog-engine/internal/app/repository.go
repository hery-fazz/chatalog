package app

import (
	"database/sql"

	"github.com/defryfazz/fazztalog/internal/merchant"
	merchantrepo "github.com/defryfazz/fazztalog/internal/merchant/repository"
)

type repository struct {
	Merchant merchant.Repository
}

func setupRepositories(db *sql.DB) repository {
	merchantRepo := merchantrepo.NewMerchantRepository(db)

	return repository{
		Merchant: merchantRepo,
	}
}
