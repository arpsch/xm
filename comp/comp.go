package comp

import (
	"context"

	"github.com/arpsch/xm/model"
	"github.com/arpsch/xm/store"
)

// CompanyApp represents the behavour on compay object
type CompanyApp interface {
	CreateCompany(ctx context.Context, c model.Company) (string, error)
	ListCompanies(ctx context.Context, q store.ListQuery) ([]model.Company, int, error)
	GetCompany(ctx context.Context, id string) (*model.Company, error)
	UpdateCompany(ctx context.Context, id string, cu model.CompanyUpdate) error
	DeleteCompany(ctx context.Context, id string) error
}

// app is an app object
type companyApp struct {
	store store.DataStore
}

// NewApp initialize a new company App
func NewApp(ds store.DataStore) (*companyApp, error) {

	app := &companyApp{
		store: ds,
	}

	return app, nil
}

func (ca *companyApp) CreateCompany(ctx context.Context, c model.Company) (string, error) {
	return ca.store.CreateCompany(ctx, c)
}

func (ca *companyApp) ListCompanies(ctx context.Context, q store.ListQuery) ([]model.Company, int, error) {
	return ca.store.ListCompanies(ctx, q)

}

func (ca *companyApp) GetCompany(ctx context.Context, id string) (*model.Company, error) {
	return ca.store.GetCompany(ctx, id)
}

func (ca *companyApp) UpdateCompany(ctx context.Context, id string, cu model.CompanyUpdate) error {
	return ca.store.UpdateCompany(ctx, id, cu)
}

func (ca *companyApp) DeleteCompany(ctx context.Context, id string) error {
	return ca.store.DeleteCompany(ctx, id)
}
