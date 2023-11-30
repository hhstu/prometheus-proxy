package prom_proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hhstu/prometheus-proxy/config"
	"github.com/hhstu/prometheus-proxy/log"
	"github.com/prometheus/prometheus/pkg/labels"
	"net/http"
)

func DoRequest(w http.ResponseWriter, req *http.Request, params map[string]string) {

	var matchers []*labels.Matcher
	for _, header := range config.AppConfig.Headers {

		value := req.Header.Get(header)
		matcher := labels.Matcher{
			Name:  header,
			Type:  labels.MatchEqual,
			Value: value,
		}
		matchers = append(matchers, &matcher)
	}

	for key, value := range params {

		matcher := labels.Matcher{
			Name:  key,
			Type:  labels.MatchEqual,
			Value: value,
		}

		matchers = append(matchers, &matcher)
	}
	e := NewEnforcer(true, matchers...)

	q, found1, err := EnforceQueryValues(e, req.URL.Query())
	if err != nil {
		switch err.(type) {
		case IllegalLabelMatcherError:
			prometheusAPIError(w, err.Error(), http.StatusBadRequest)
		case QueryParseError:
			prometheusAPIError(w, err.Error(), http.StatusBadRequest)
		case EnforceLabelError:
			prometheusAPIError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	req.URL.RawQuery = q

	var found2 bool
	// Enforce the query in the POST body if needed.
	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			prometheusAPIError(w, err.Error(), http.StatusBadRequest)
		}
		q, found2, err = EnforceQueryValues(e, req.PostForm)
		if err != nil {
			switch err.(type) {
			case IllegalLabelMatcherError:
				prometheusAPIError(w, err.Error(), http.StatusBadRequest)
			case QueryParseError:
				prometheusAPIError(w, err.Error(), http.StatusBadRequest)
			case EnforceLabelError:
				prometheusAPIError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// If no query was found, return early.
	if !found1 && !found2 {
		return
	}
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		prometheusAPIError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()
	bufio.NewReader(resp.Body).WriteTo(w)
	return

}

func prometheusAPIError(w http.ResponseWriter, errorMessage string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	res := map[string]string{"status": "error", "errorType": "prom-label-proxy", "error": errorMessage}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Logger.Error(fmt.Sprintf("error: Failed to encode json: %v", err))
	}
}
