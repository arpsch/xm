package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// Company represents the company information
type Company struct {
	ID string `json:"id" bson:"_id,omitempty"`

	Name    string `json:"name" bson:"name,omitempty"`
	Code    string `json:"code" bson:"code,omitempty"`
	Country string `jsoon:"country" bson:"country,omitempty"`
	Website string `json:"website" bson:"website,omitempty"`
	Phone   string `json:"phone" bson:"phone,omitmepty"`

	CreatedTs time.Time `json:"crated_ts" bson:"created_ts,omitempty"`
	UpdatedTs time.Time `json:"updated_ts" bson:"updated_ts,omitempty"`
}

func (comp Company) Validate() error {
	err := validation.ValidateStruct(&comp,
		validation.Field(&comp.Name, validation.Required),
		validation.Field(&comp.Country, validation.Required),
		validation.Field(&comp.Code, validation.Required, is.CountryCode2),
		validation.Field(&comp.Phone, is.E164),
		validation.Field(&comp.Website, is.URL),
	)
	if err != nil {
		return err
	}
	return nil
}

// CompanyUpdate allows updating the company information
type CompanyUpdate struct {
	Website string `json:"website" bson:"website,omitempty"`
	Phone   string `json:"phone" bson:"phone,omitmepty"`

	UpdatedTs time.Time `json:"updated_ts" bson:"updated_ts,omitempty"`
}

func (compUp CompanyUpdate) Validate() error {
	err := validation.ValidateStruct(&compUp,
		validation.Field(&compUp.Phone, is.E164),
		validation.Field(&compUp.Website, is.URL),
	)
	if err != nil {
		return err
	}
	return nil
}
