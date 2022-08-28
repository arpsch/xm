package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	ipapi "github.com/arpsch/xm/client"
	"github.com/arpsch/xm/comp"
	"github.com/arpsch/xm/model"
	"github.com/arpsch/xm/store"
	"github.com/arpsch/xm/utils"
	"github.com/pkg/errors"

	"github.com/julienschmidt/httprouter"
)

type ApiHandler struct {
	App comp.CompanyApp
}

func NewApiHandler(app comp.CompanyApp) *ApiHandler {
	return &ApiHandler{
		App: app,
	}
}

func NewRouter(app comp.CompanyApp) *httprouter.Router {
	apiHandler := NewApiHandler(app)

	router := httprouter.New()
	router.HandlerFunc("GET", "/api/v1/companies", apiHandler.ListCompaniesHandler)
	router.HandlerFunc("GET", "/api/v1/companies/:id", apiHandler.GetCompanyHandler)

	router.HandlerFunc("POST", "/api/v1/companies", apiHandler.CreateCompanyHandler)
	router.HandlerFunc("PUT", "/api/v1/companies/:id", apiHandler.UpdateCompanyHandler)
	router.HandlerFunc("DELETE", "/api/v1/companies/:id", apiHandler.DeleteCompanyHandler)

	return router
}

const (
	queryParamGroup          = "group"
	queryParamSort           = "sort"
	queryParamValueSeparator = ":"
	sortOrderAsc             = "asc"
	sortOrderDesc            = "desc"
	sortAttributeNameIdx     = 0
	sortOrderIdx             = 1
	filterEqOperatorIdx      = 0
)

const (
	// CYPRUS_CC is the 2-letter Cyprus country code
	CYPRUS_CC = "CY"
)

func parseCompany(r *http.Request) (model.Company, error) {
	comp := model.Company{}

	//decode body
	err := json.NewDecoder(r.Body).Decode(&comp)
	if err != nil {
		return model.Company{}, errors.Wrap(err, "failed to decode request body")
	}

	if err := comp.Validate(); err != nil {
		return model.Company{}, err
	}

	return comp, nil
}

func parseCompanyUpdate(r *http.Request) (model.CompanyUpdate, error) {
	compUp := model.CompanyUpdate{}

	//decode body
	err := json.NewDecoder(r.Body).Decode(&compUp)
	if err != nil {
		return model.CompanyUpdate{}, errors.Wrap(err, "failed to decode request body")
	}

	if err := compUp.Validate(); err != nil {
		return model.CompanyUpdate{}, err
	}

	return compUp, nil
}

func parseFilterParams(r *http.Request) ([]store.Filter, error) {
	knownParams := []string{utils.PageName, utils.PerPageName}
	filters := make([]store.Filter, 0)
	var filter store.Filter
	for name := range r.URL.Query() {
		if utils.ContainsString(name, knownParams) {
			continue
		}
		valueStr, err := utils.ParseQueryParmStr(r, name, false, nil)
		if err != nil {
			return nil, err
		}
		valueStrArray := strings.Split(valueStr, queryParamValueSeparator)
		filter = store.Filter{AttrName: name}
		if len(valueStrArray) == 2 {
			switch valueStrArray[filterEqOperatorIdx] {
			case "eq":
				filter.Operator = store.Eq
			default:
				return nil, errors.New("invalid filter operator")
			}
			filter.Value = valueStrArray[filterEqOperatorIdx+1]
		} else {
			filter.Operator = store.Eq
			filter.Value = valueStr
		}
		floatValue, err := strconv.ParseFloat(filter.Value, 64)
		if err == nil {
			filter.ValueFloat = &floatValue
		}

		filters = append(filters, filter)
	}
	return filters, nil
}

func validateClientOriginCountry(r *http.Request) (bool, error) {
	remoteIP, err := utils.RetrievRemoteIP(r)
	if err != nil {
		return false, err
	}

	country, err := ipapi.IPAPI_GetCountryNameByIP(r.Context(), remoteIP)
	if err != nil {
		return false, err
	}

	if country != CYPRUS_CC {
		return false, err
	}
	return true, nil
}

// CreateCompanyHandler allows to create a company entry in the DB if the client
// is making request from Cyprus location.
// Return the entry with ID param added
func (ah *ApiHandler) CreateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	b, err := validateClientOriginCountry(r)
	if err != nil {
		http.Error(w, "failed to retrieve client country: "+err.Error(), http.StatusForbidden)
		return
	}

	if b == false {
		http.Error(w, "you're not authorized", http.StatusForbidden)
		return
	}

	comp, err := parseCompany(r)
	if err != nil {
		http.Error(w, "failed to parse the payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := ah.App.CreateCompany(ctx, comp)
	if err != nil {
		if errors.Is(store.ErrCompanyExists, err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to create the company entry: "+err.Error(), http.StatusInternalServerError)
		return
	}
	comp.ID = id
	cJ, err := json.Marshal(&comp)
	if err != nil {
		http.Error(w, "internal server error in creating the db", http.StatusInternalServerError)
		return
	}

	w.Write(cJ)
}

// ListCompaniesHandler fetches the all companies in the db.
// Entries could be fetched based on the filters if set
// Returns array of companies
func (ah *ApiHandler) ListCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, perPage, err := utils.ParsePagination(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filters, err := parseFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ld := store.ListQuery{Skip: int((page - 1) * perPage),
		Limit:   int(perPage),
		Filters: filters,
	}

	companies, totalCount, err := ah.App.ListCompanies(ctx, ld)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasNext := totalCount > int(page*perPage)
	links := utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add("Link", l)
	}

	cJ, err := json.Marshal(&companies)
	if err != nil {
		http.Error(w, "internal server error in retrieving companies", http.StatusInternalServerError)
		return
	}

	w.Write(cJ)
}

// GetCompanyHandler fetches a particular company information based on
// supplied company id
func (ah *ApiHandler) GetCompanyHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	id := utils.ParsePathParamId(r)
	if id == "" {
		http.Error(w, "id path param is empty", http.StatusBadRequest)
		return
	}

	comp, err := ah.App.GetCompany(ctx, id)
	if err != nil {
		if errors.Is(store.ErrCompanyNotFound, err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error in retrieving the company: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	cJ, err := json.Marshal(&comp)
	if err != nil {
		http.Error(w, "internal server error in retrieving company: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.Write(cJ)
}

// UpdateCompanyHandler updates the allowed fields for a selected company by its id
func (ah *ApiHandler) UpdateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := utils.ParsePathParamId(r)
	if id == "" {
		http.Error(w, "id path param is empty", http.StatusBadRequest)
		return
	}

	compUp, err := parseCompanyUpdate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = ah.App.UpdateCompany(ctx, id, compUp)
	if err != nil {
		if errors.Is(store.ErrCompanyNotFound, err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error in updating the company: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// DeleteCompanyHandler deletes a company information by the given id.
// But the caller must be calling from Cyprus to delete a company information
func (ah *ApiHandler) DeleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	b, err := validateClientOriginCountry(r)
	if err != nil {
		http.Error(w, "failed to retrieve client country: "+err.Error(), http.StatusForbidden)
		return
	}

	if b == false {
		http.Error(w, "you're not authorized", http.StatusForbidden)
		return
	}

	id := utils.ParsePathParamId(r)
	if id == "" {
		http.Error(w, "id path param is empty", http.StatusBadRequest)
		return
	}

	err = ah.App.DeleteCompany(ctx, id)
	if err != nil {
		http.Error(w, "internal server error deleting the company: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
