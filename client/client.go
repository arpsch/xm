package client

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// IPAPIURL https://ipapi.co/api/#complete-location
	IPAPIURL = "https://ipapi.co/:ip/country/"

	// default request timeout, 10s
	defaultReqTimeout = time.Duration(10) * time.Second
)

func IPAPI_GetCountryNameByIP(ctx context.Context, ip string) (string, error) {

	if ip == "" {
		return "", errors.New("missing IP")
	}

	if r := net.ParseIP(ip); r == nil {
		return "", errors.New("Invalid IP")
	}

	repl := strings.NewReplacer(":ip", ip)
	uri := repl.Replace(IPAPIURL)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for GET /:ip/country/")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultReqTimeout)
	defer cancel()

	client := http.Client{}
	rsp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "", errors.Wrap(err, "GET /:ip/country/ request failed")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", errors.Wrap(err, "GET  /:ip/country/  request failed")
	}

	country, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(country), nil
}
