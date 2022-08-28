package store

import (
	"context"
	"errors"

	"github.com/arpsch/xm/model"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
	ErrCompanyExists   = errors.New("company exists")
)

//  DataStore  represents behavour on DataStore
type DataStore interface {
	CreateCompany(ctx context.Context, c model.Company) (string, error)
	ListCompanies(ctx context.Context, q ListQuery) ([]model.Company, int, error)
	GetCompany(ctx context.Context, id string) (*model.Company, error)
	UpdateCompany(ctx context.Context, id string, cu model.CompanyUpdate) error
	DeleteCompany(ctx context.Context, id string) error
}
