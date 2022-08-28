package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	api_http "github.com/arpsch/xm/api/http"
	"github.com/arpsch/xm/model"
	"github.com/arpsch/xm/store/mongo"
	"github.com/pkg/errors"
)

var ah *api_http.ApiHandler
var ds *mongo.MongoStore

func testSetup() error {
	var err error

	mgoUrl, err := url.Parse("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	storeConfig := mongo.MongoStoreConfig{
		MongoURL: mgoUrl,
		DbName:   "xm",
	}
	ds, err = mongo.NewMongoStore(context.Background(), storeConfig)
	if err != nil {
		return err
	}
	ah = api_http.NewApiHandler(ds)

	return nil
}

func testTeardown() error {
	err := ds.DropDatabase(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	if err := testSetup(); err != nil {
		os.Exit(-1)
	}

	exitCode := m.Run()

	if err := testTeardown(); err != nil {
		os.Exit(-1)
	}

	// Exit
	os.Exit(exitCode)
}

func TestCreateCompany(t *testing.T) {

	tt := []struct {
		name       string
		method     string
		URL        string
		input      model.Company
		want       model.Company
		statusCode int
		err        error
		OriginIP   string
	}{
		{
			name:   "with a valid company information",
			method: http.MethodPost,
			URL:    "/api/v1/companies",
			input: model.Company{
				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},

			want: model.Company{
				ID: "1234566",
			},
			err:      nil,
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusOK,
		},
		{
			name:   "create a company outside Cyprus",
			method: http.MethodPost,
			URL:    "/api/v1/companies",
			input: model.Company{
				ID:      "1234566",
				Name:    "jio",
				Country: "India",
				Code:    "IN",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},

			want:       model.Company{},
			err:        errors.New("you're not authorized"),
			statusCode: http.StatusForbidden,
			OriginIP:   "127.0.0.1:8080",
		},

		{
			name:   "adding the existing company",
			method: http.MethodPost,
			URL:    "/api/v1/companies",
			input: model.Company{
				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},

			err: errors.New("company exists"),

			statusCode: http.StatusBadRequest,

			want:     model.Company{},
			OriginIP: "185.193.151.255:8080",
		},
		{
			name:   "without name",
			method: http.MethodPost,
			URL:    "/api/v1/companies",
			input: model.Company{
				ID:      "1234566",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},

			err: errors.New("name field required"),

			statusCode: http.StatusBadRequest,

			want:     model.Company{},
			OriginIP: "185.193.151.255:8080",
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			compJson, _ := json.Marshal(&tc.input)
			req, err := http.NewRequest(
				tc.method,
				tc.URL,
				bytes.NewBuffer(compJson),
			)
			// set the client IP
			req.RemoteAddr = tc.OriginIP

			if err != nil {
				t.Fatalf("Could not create a request %v", err)
			}

			if tc.name == "adding the existing company" {
				// this is to add the same document befor the test
				recPrep := httptest.NewRecorder()
				reqPrep, err := http.NewRequest(
					tc.method,
					tc.URL,
					bytes.NewBuffer(compJson),
				)
				reqPrep.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a prep request %v", err)
				}

				ah.CreateCompanyHandler(recPrep, reqPrep)
			}

			ah.CreateCompanyHandler(rec, req)

			if rec.Code != tc.statusCode {
				b, _ := ioutil.ReadAll(rec.Body)
				t.Logf("error message: %s\n", b)
				t.Errorf("expected status %v, received status %v", tc.statusCode, rec.Code)
			}

			if err := ds.DropDatabase(context.Background()); err != nil {
				t.Errorf("database clean-up failed: %s\n", err)
			}
		})
	}

}

func TestGetCompanies(t *testing.T) {

	tt := []struct {
		name        string
		method      string
		URL         string
		preExisting []model.Company
		want        []model.Company
		statusCode  int
		err         error
		OriginIP    string
	}{
		{
			name:   "with 1 company information",
			method: http.MethodGet,
			URL:    "/api/v1/companies",
			preExisting: []model.Company{
				{
					ID:      "1234566",
					Name:    "Airtel",
					Country: "Cyprus",
					Code:    "CY",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
			},
			want: []model.Company{
				{
					ID:      "1234566",
					Name:    "Airtel",
					Country: "Cyprus",
					Code:    "CY",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
			},
			err:      nil,
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusOK,
		},
		{
			name:        "with no companies information",
			method:      http.MethodGet,
			URL:         "/api/v1/companies",
			preExisting: []model.Company{},
			want:        []model.Company{},
			err:         nil,
			OriginIP:    "127.0.0.1:8080",
			statusCode:  http.StatusOK,
		},
		{
			name:   "filter company information by name",
			method: http.MethodGet,
			URL:    "/api/v1/companies?name=vodafone",
			preExisting: []model.Company{
				{
					ID:      "1234566",
					Name:    "Airtel",
					Country: "Cyprus",
					Code:    "CY",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
				{
					ID:      "1234577",
					Name:    "vodafone",
					Country: "Cyprus",
					Code:    "CY",
					Website: "vodafone.cy",
					Phone:   "+35722111199",
				},
			},
			want: []model.Company{
				{
					ID:      "1234577",
					Name:    "vodafone",
					Country: "Cyprus",
					Code:    "CY",
					Website: "vodafone.cy",
					Phone:   "+35722111199",
				},
			},
			err:      nil,
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusOK,
		},
		{
			name:   "filter company information by country code",
			method: http.MethodGet,
			URL:    "/api/v1/companies?code=CY",
			preExisting: []model.Company{
				{
					ID:      "1234566",
					Name:    "Airtel",
					Country: "Cyprus",
					Code:    "CY",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
				{
					ID:      "1234577",
					Name:    "vodafone",
					Country: "Cyprus",
					Code:    "CY",
					Website: "vodafone.cy",
					Phone:   "+35722111199",
				},
			},
			want: []model.Company{
				{
					ID:      "1234566",
					Name:    "Airtel",
					Country: "Cyprus",
					Code:    "CY",
					Website: "airtel.cy",
					Phone:   "+35722111111",
				},
				{
					ID:      "1234577",
					Name:    "vodafone",
					Country: "Cyprus",
					Code:    "CY",
					Website: "vodafone.cy",
					Phone:   "+35722111199",
				},
			},
			err:      nil,
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(
				tc.method,
				tc.URL,
				nil,
			)
			if err != nil {
				t.Fatalf("Could not create a get request %v", err)
			}

			// this is to add the same document befor the test
			for _, comp := range tc.preExisting {
				compJson, _ := json.Marshal(&comp)
				recPrep := httptest.NewRecorder()
				reqPrep, err := http.NewRequest(
					http.MethodPost,
					tc.URL,
					bytes.NewBuffer(compJson),
				)
				reqPrep.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a prep request %v", err)
				}

				ah.CreateCompanyHandler(recPrep, reqPrep)
			}

			ah.ListCompaniesHandler(rec, req)

			if rec.Code != tc.statusCode {
				b, _ := ioutil.ReadAll(rec.Body)

				t.Logf("error message: %s\n", b)
				t.Errorf("expected status %v, received status %v", tc.statusCode, rec.Code)
			} else {
				companies := []model.Company{}
				err := json.NewDecoder(rec.Body).Decode(&companies)
				if err != nil {
					t.Errorf("failed to decode response body: %v", err)
				}
				if len(tc.want) != len(companies) {
					t.Errorf("expected count %v, received count %v: %v", len(tc.want), len(companies), companies)
				}
			}

			// clean-up the db before next TC
			if err := ds.DropDatabase(context.Background()); err != nil {
				t.Fatalf("database clean-up failed: %s\n", err)
			}
		})
	}
}

func TestGetCompany(t *testing.T) {

	tt := []struct {
		name       string
		method     string
		URL        string
		want       model.Company
		statusCode int
		OriginIP   string
	}{
		{
			name:   "with a valid company information",
			method: http.MethodGet,
			URL:    "/api/v1/companies/%s",
			want: model.Company{

				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusOK,
		},
		{
			name:       "with no companies information",
			method:     http.MethodGet,
			URL:        "/api/v1/companies/%s",
			want:       model.Company{ID: "1234566"},
			OriginIP:   "127.0.0.1:8080",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(
				tc.method,
				fmt.Sprintf(tc.URL, tc.want.ID),
				nil,
			)
			if err != nil {
				t.Fatalf("Could not create a get request %v", err)
			}

			if tc.name == "with a valid company information" {
				// this is to add the same document befor the test
				compJson, _ := json.Marshal(&tc.want)
				recPrep := httptest.NewRecorder()
				reqPrep, err := http.NewRequest(
					http.MethodPost,
					"/api/v1/companies",
					bytes.NewBuffer(compJson),
				)
				reqPrep.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a prep request %v", err)
				}

				ah.CreateCompanyHandler(recPrep, reqPrep)
			}

			ah.ListCompaniesHandler(rec, req)

			if rec.Code != tc.statusCode {
				b, _ := ioutil.ReadAll(rec.Body)
				t.Logf("error message: %s\n", b)
				t.Errorf("expected status %v, received status %v", tc.statusCode, rec.Code)
			}

			if err := ds.DropDatabase(context.Background()); err != nil {
				t.Fatalf("database clean-up failed: %s\n", err)
			}

		})
	}
}

func TestUpdateCompany(t *testing.T) {

	tt := []struct {
		name       string
		method     string
		URL        string
		input      model.CompanyUpdate
		existing   model.Company
		want       model.Company
		statusCode int
		err        error
		OriginIP   string
	}{
		{
			name:   "with a valid company information",
			method: http.MethodPut,
			URL:    "/api/v1/companies/%s",
			input: model.CompanyUpdate{

				Website: "airtelin.cy",
				Phone:   "+35722111122",
			},

			existing: model.Company{

				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},
			want: model.Company{

				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtelin.cy",
				Phone:   "+35722111122",
			},

			err:      nil,
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusAccepted,
		},
		{
			name:       "with no companies information",
			method:     http.MethodPut,
			URL:        "/api/v1/companies/%s",
			want:       model.Company{ID: "1234566"},
			err:        errors.New("company not found"),
			OriginIP:   "127.0.0.1:8080",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			compJson, _ := json.Marshal(&tc.want)
			req, err := http.NewRequest(
				tc.method,
				fmt.Sprintf(tc.URL, tc.want.ID),
				bytes.NewBuffer(compJson),
			)
			if err != nil {
				t.Fatalf("Could not create a put  request %v", err)
			}

			if tc.name == "with a valid company information" {
				// this is to add the same document befor the test
				compJson, _ := json.Marshal(&tc.existing)
				recPrep := httptest.NewRecorder()
				reqPrep, err := http.NewRequest(
					http.MethodPost,
					"/api/v1/companies",
					bytes.NewBuffer(compJson),
				)
				reqPrep.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a prep request %v", err)
				}

				ah.CreateCompanyHandler(recPrep, reqPrep)
			}

			ah.UpdateCompanyHandler(rec, req)

			if rec.Code != tc.statusCode {
				b, _ := ioutil.ReadAll(rec.Body)
				t.Logf("error message: %s\n", b)
				t.Errorf("expected status %v, received status %v", tc.statusCode, rec.Code)
			} else if rec.Code != http.StatusBadRequest {
				recPost := httptest.NewRecorder()
				reqPost, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf(tc.URL, tc.want.ID),
					nil,
				)
				reqPost.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a post request %v", err)
				}

				ah.GetCompanyHandler(recPost, reqPost)

				company := model.Company{}
				err = json.NewDecoder(recPost.Body).Decode(&company)
				if err != nil {
					t.Errorf("failed to decode response body: %v", err)
				}
				if company.Phone != tc.want.Phone || company.Website != tc.want.Website {
					t.Errorf("expected count %v, received count %v", tc.want, company)
				}

			}

			if err := ds.DropDatabase(context.Background()); err != nil {
				t.Fatalf("database clean-up failed: %s\n", err)
			}
		})
	}
}

func TestDeleteCompany(t *testing.T) {

	tt := []struct {
		name       string
		method     string
		URL        string
		existing   model.Company
		statusCode int
		OriginIP   string
	}{
		{
			name:   "with a valid company information",
			method: http.MethodDelete,
			URL:    "/api/v1/companies/%s",
			existing: model.Company{

				ID:      "1234566",
				Name:    "Airtel",
				Country: "Cyprus",
				Code:    "CY",
				Website: "airtel.cy",
				Phone:   "+35722111111",
			},
			OriginIP: "185.193.151.255:8080",

			statusCode: http.StatusNoContent,
		},
		{
			name:       "with no companies information",
			method:     http.MethodPut,
			URL:        "/api/v1/companies/%s",
			OriginIP:   "185.193.151.255:8080",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "deleting from outside cyprus",
			method:     http.MethodPut,
			URL:        "/api/v1/companies/%s",
			OriginIP:   "127.0.0.1:8080",
			statusCode: http.StatusForbidden,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(
				tc.method,
				fmt.Sprintf(tc.URL, tc.existing.ID),
				nil,
			)
			if err != nil {
				t.Fatalf("Could not create a delete request %v", err)
			}

			if tc.name == "with a valid company information" {
				// this is to add the same document befor the test
				compJson, _ := json.Marshal(&tc.existing)
				recPrep := httptest.NewRecorder()
				reqPrep, err := http.NewRequest(
					http.MethodPost,
					"/api/v1/companies",
					bytes.NewBuffer(compJson),
				)
				reqPrep.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a prep request %v", err)
				}

				ah.CreateCompanyHandler(recPrep, reqPrep)
			}

			req.RemoteAddr = tc.OriginIP
			ah.DeleteCompanyHandler(rec, req)

			if rec.Code != tc.statusCode {
				b, _ := ioutil.ReadAll(rec.Body)
				t.Logf("error message: %s\n", b)
				t.Errorf("expected status %v, received status %v", tc.statusCode, rec.Code)
			} else if rec.Code != http.StatusBadRequest && rec.Code != http.StatusForbidden {
				recPost := httptest.NewRecorder()
				reqPost, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf(tc.URL, tc.existing.ID),
					nil,
				)
				reqPost.RemoteAddr = tc.OriginIP
				if err != nil {
					t.Fatalf("Could not create a post request %v", err)
				}

				ah.GetCompanyHandler(recPost, reqPost)

				if recPost.Code != http.StatusBadRequest {
					t.Errorf("failed to delete the company")
				}
			}

			if err := ds.DropDatabase(context.Background()); err != nil {
				t.Fatalf("database clean-up failed: %s\n", err)
			}

		})
	}
}
