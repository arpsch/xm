package mongo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arpsch/xm/model"
	"github.com/arpsch/xm/store"
	. "github.com/arpsch/xm/store/mongo"
)

// test funcs
func TestMongoGetCompanies(t *testing.T) {

	inputCompanies := []model.Company{
		{
			ID:      "1234566",
			Name:    "Airtel",
			Country: "Cyprus",
			Code:    "CY",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
		{
			ID:      "12345689",
			Name:    "jio",
			Country: "India",
			Code:    "IN",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
	}

	testCases := map[string]struct {
		expected  []model.Company
		compTotal int
		skip      int
		limit     int
		filters   []store.Filter
		sort      *store.Sort
	}{
		"all companies, no skip, no limit": {
			expected:  inputCompanies,
			compTotal: len(inputCompanies),
			skip:      0,
			limit:     20,
			filters:   nil,
			sort:      nil,
		},
		"filter on attribute (equal attribute)": {
			expected:  []model.Company{inputCompanies[1]},
			compTotal: 1,
			skip:      0,
			limit:     20,
			filters: []store.Filter{
				{
					AttrName: "name",
					Value:    "Airtel", Operator: store.Eq,
				},
			},
			sort: nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		ctx := context.Background()

		for _, d := range inputCompanies {
			c := ds.Database(ctx).Collection(DbCompaniesColl)
			_, err := c.InsertOne(ctx, d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		companies, totalCount, err := ds.ListCompanies(ctx,
			store.ListQuery{
				Skip:    tc.skip,
				Limit:   tc.limit,
				Filters: tc.filters,
				Sort:    tc.sort})
		assert.NoError(t, err, "failed to get companies")

		assert.Equal(t, len(tc.expected), len(companies))
		assert.Equal(t, tc.compTotal, totalCount)

		err = ds.DropDatabase(ctx)
		assert.NoError(t, err, "failed to clean companies db")
	}
}

func TestMongoGetCompany(t *testing.T) {

	inputCompanies := []model.Company{
		{
			ID:      "1234566",
			Name:    "Airtel",
			Country: "Cyprus",
			Code:    "CY",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
		{
			ID:      "12345689",
			Name:    "jio",
			Country: "India",
			Code:    "IN",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
	}

	testCases := map[string]struct {
		expected  model.Company
		compTotal int
		skip      int
		limit     int
		filters   []store.Filter
		sort      *store.Sort
	}{
		"one company by id": {
			expected:  inputCompanies[1],
			compTotal: len(inputCompanies),
			skip:      0,
			limit:     20,
			filters:   nil,
			sort:      nil,
		},
	}

	for name := range testCases {
		t.Logf("test case: %s", name)
		ctx := context.Background()

		for _, d := range inputCompanies {
			c := ds.Database(ctx).Collection(DbCompaniesColl)
			_, err := c.InsertOne(ctx, d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		company, err := ds.GetCompany(ctx, inputCompanies[1].ID)
		assert.NoError(t, err, "failed to get company")

		assert.Equal(t, company.ID, inputCompanies[1].ID)

		err = ds.DropDatabase(ctx)
		assert.NoError(t, err, "failed to clean companies db")
	}
}

func TestMongoCreateCompany(t *testing.T) {

	inputCompanies := []model.Company{
		{
			ID:      "1234566",
			Name:    "Airtel",
			Country: "Cyprus",
			Code:    "CY",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
		{
			ID:      "12345689",
			Name:    "jio",
			Country: "India",
			Code:    "IN",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
	}

	testCases := map[string]struct {
		expected  []model.Company
		compTotal int
		skip      int
		limit     int
		filters   []store.Filter
		sort      *store.Sort
	}{
		"one company by id": {
			expected:  inputCompanies,
			compTotal: len(inputCompanies),
			skip:      0,
			limit:     20,
			filters:   nil,
			sort:      nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		ctx := context.Background()

		for _, d := range inputCompanies {
			c := ds.Database(ctx).Collection(DbCompaniesColl)
			_, err := c.InsertOne(ctx, d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		companies, totalCount, err := ds.ListCompanies(ctx,
			store.ListQuery{
				Skip:    tc.skip,
				Limit:   tc.limit,
				Filters: tc.filters,
				Sort:    tc.sort})
		assert.NoError(t, err, "failed to get companies")

		assert.Equal(t, len(tc.expected), len(companies))
		assert.Equal(t, tc.compTotal, totalCount)

		err = ds.DropDatabase(ctx)
		assert.NoError(t, err, "failed to clean companies db")
	}
}

func TestMongoUpdateCompany(t *testing.T) {

	inputCompanies := []model.Company{
		{
			ID:      "1234566",
			Name:    "Airtel",
			Country: "Cyprus",
			Code:    "CY",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
		{
			ID:      "12345689",
			Name:    "jio",
			Country: "India",
			Code:    "IN",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
	}

	testCases := map[string]struct {
		expected  model.Company
		input     model.CompanyUpdate
		compTotal int
		skip      int
		limit     int
		filters   []store.Filter
		sort      *store.Sort
	}{
		"update company by id": {
			input: model.CompanyUpdate{
				Website: "jio",
				Phone:   "+35722111777",
			},
			expected: model.Company{
				ID:      "12345689",
				Name:    "jio",
				Country: "India",
				Code:    "IN",
				Website: "jio",
				Phone:   "+35722111777",
			},
			compTotal: len(inputCompanies),
			skip:      0,
			limit:     20,
			filters:   nil,
			sort:      nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		ctx := context.Background()

		for _, d := range inputCompanies {
			c := ds.Database(ctx).Collection(DbCompaniesColl)
			_, err := c.InsertOne(ctx, d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		err := ds.UpdateCompany(ctx, tc.expected.ID, tc.input)
		assert.NoError(t, err, "failed to update company")

		comp, err := ds.GetCompany(ctx, tc.expected.ID)
		assert.NoError(t, err, "failed to get company")

		assert.Equal(t, tc.expected.Website, comp.Website)
		assert.Equal(t, tc.expected.Phone, comp.Phone)

		err = ds.DropDatabase(ctx)
		assert.NoError(t, err, "failed to clean companies db")
	}
}

func TestMongoDeleteCompany(t *testing.T) {

	inputCompanies := []model.Company{
		{
			ID:      "1234566",
			Name:    "Airtel",
			Country: "Cyprus",
			Code:    "CY",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
		{
			ID:      "12345689",
			Name:    "jio",
			Country: "India",
			Code:    "IN",
			Website: "airtel.cy",
			Phone:   "+35722111111",
		},
	}

	testCases := map[string]struct {
		expected  []model.Company
		input     string
		compTotal int
	}{
		"update company by id": {
			input: "1234566",
			expected: []model.Company{
				{
					ID:      "12345689",
					Name:    "jio",
					Country: "India",
					Code:    "IN",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
			},
			compTotal: len(inputCompanies) - 1,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		ctx := context.Background()

		for _, d := range inputCompanies {
			c := ds.Database(ctx).Collection(DbCompaniesColl)
			_, err := c.InsertOne(ctx, d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		err := ds.DeleteCompany(ctx, tc.input)
		assert.NoError(t, err, "failed to delete company")

		err = ds.DeleteCompany(ctx, tc.input)
		assert.Error(t, err, store.ErrCompanyNotFound)

		err = ds.DropDatabase(ctx)
		assert.NoError(t, err, "failed to clean companies db")
	}
}
