package utils

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//pagination constants
const (
	PageName       = "page"
	PerPageName    = "per_page"
	PageMin        = 1
	PageDefault    = 1
	PerPageMin     = 1
	PerPageMax     = 500
	PerPageDefault = 20
	LinkHdr        = "Link"
	LinkTmpl       = "<%s?%s>; rel=\"%s\""
	LinkPrev       = "prev"
	LinkNext       = "next"
	LinkFirst      = "first"
	DefaultScheme  = "http"
)

const (
	URLPrefix = "/api/v1/companies/"
)

//error msgs
func MsgQueryParmInvalid(name string) string {
	return fmt.Sprintf("Can't parse param %s", name)
}

func MsgQueryParmMissing(name string) string {
	return fmt.Sprintf("Missing required param %s", name)
}

func MsgQueryParmLimit(name string) string {
	return fmt.Sprintf("Param %s is out of bounds", name)
}

func MsgQueryParmOneOf(name string, allowed []string) string {
	return fmt.Sprintf("Param %s must be one of %v", name, allowed)
}

//query param parsing/validation
func ParseQueryParmUInt(r *http.Request, name string, required bool, min, max, def uint64) (uint64, error) {
	strVal := r.URL.Query().Get(name)

	if strVal == "" {
		if required {
			return 0, errors.New(MsgQueryParmMissing(name))
		} else {
			return def, nil
		}
	}

	uintVal, err := strconv.ParseUint(strVal, 10, 32)
	if err != nil {
		return 0, errors.New(MsgQueryParmInvalid(name))
	}

	if uintVal < min || uintVal > max {
		return 0, errors.New(MsgQueryParmLimit(name))
	}

	return uintVal, nil
}

func ParseQueryParmBool(r *http.Request, name string, required bool, def *bool) (*bool, error) {
	strVal := r.URL.Query().Get(name)

	if strVal == "" {
		if required {
			return nil, errors.New(MsgQueryParmMissing(name))
		} else {
			return def, nil
		}
	}

	boolVal, err := strconv.ParseBool(strVal)
	if err != nil {
		return nil, errors.New(MsgQueryParmInvalid(name))
	}

	return &boolVal, nil
}

func ParseQueryParmStr(r *http.Request, name string, required bool, allowed []string) (string, error) {
	val := r.URL.Query().Get(name)

	if val == "" {
		if required {
			return "", errors.New(MsgQueryParmMissing(name))
		}
	} else {
		if allowed != nil && !ContainsString(val, allowed) {
			return "", errors.New(MsgQueryParmOneOf(name, allowed))
		}
	}

	val, err := url.QueryUnescape(val)
	if err != nil {
		return "", errors.New(MsgQueryParmInvalid(name))
	}

	return val, nil
}

//pagination helpers
func ParsePagination(r *http.Request) (uint64, uint64, error) {
	page, err := ParseQueryParmUInt(r, PageName, false, PageMin, math.MaxUint64, PageDefault)
	if err != nil {
		return 0, 0, err
	}

	per_page, err := ParseQueryParmUInt(r, PerPageName, false, PerPageMin, PerPageMax, PerPageDefault)
	if err != nil {
		return 0, 0, err
	}

	return page, per_page, nil
}

func MakePageLinkHdrs(r *http.Request, page, per_page uint64, has_next bool) []string {
	var links []string

	pathitems := strings.Split(r.URL.Path, "/")
	resource := pathitems[len(pathitems)-1]
	query := r.URL.Query()

	if page > 1 {
		links = append(links, MakeLink(LinkPrev, resource, query, page-1, per_page))
	}

	if has_next {
		links = append(links, MakeLink(LinkNext, resource, query, page+1, per_page))
	}

	links = append(links, MakeLink(LinkFirst, resource, query, 1, per_page))
	return links
}

func MakeLink(link_type string, resource string, query url.Values, page, per_page uint64) string {
	query.Set(PageName, strconv.Itoa(int(page)))
	query.Set(PerPageName, strconv.Itoa(int(per_page)))

	return fmt.Sprintf(LinkTmpl, resource, query.Encode(), link_type)
}

func ParsePathParamId(r *http.Request) string {
	return strings.TrimPrefix(r.URL.Path, URLPrefix)
}

func RetrievRemoteIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}
