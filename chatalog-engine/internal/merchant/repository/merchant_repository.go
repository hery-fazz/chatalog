package repository

import (
	"context"
	"database/sql"

	"github.com/defryfazz/fazztalog/internal/merchant"
)

type MerchantRepository struct {
	db *sql.DB
}

func NewMerchantRepository(db *sql.DB) *MerchantRepository {
	return &MerchantRepository{
		db: db,
	}
}

func (r *MerchantRepository) GetMerchantByPhone(ctx context.Context, phone string) (*merchant.Merchant, error) {
	query := `
		SELECT id, name, phone
		FROM merchants
		WHERE phone = ?
	`
	var res merchant.Merchant
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&res.ID,
		&res.Name,
		&res.Phone,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &res, nil
}

func (r *MerchantRepository) GetProductsByMerchantID(ctx context.Context, merchantID string) ([]merchant.Product, error) {
	query := `
		SELECT id, merchant_id, name, price
		FROM products
		WHERE merchant_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []merchant.Product
	for rows.Next() {
		var p merchant.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
