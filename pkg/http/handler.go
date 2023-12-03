package http_metrics_api

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"

	ktlogging "github.com/keytiles/lib-logging-golang"
	"github.com/keytiles/prometheus-metrics-filter/pkg/conf"
	"github.com/keytiles/prometheus-metrics-filter/pkg/rules"
)

const (
	QUERYPARAM_PROXYRULE         = "proxyRule"
	QUERYPARAM_METRICS_FETCH_URL = "metricsFetchUrl"
)

var compiledRules map[string]rules.ILineEvaluationRule

func Init(rules map[string]rules.ILineEvaluationRule) {
	compiledRules = rules
}

func printRequestInfo(LOG *ktlogging.Logger, req *http.Request) {
	LOG.Debug("incoming request %v - %v?%v", req.Method, req.URL.Path, req.URL.RawQuery)
}

func Api_ExecuteMetricsRequest(respw http.ResponseWriter, req *http.Request) {
	LOG := ktlogging.GetLogger("http.handler.Api_ExecuteMetricsRequest")

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		LOG.Debug("request took %v millis", duration)
	}()

	printRequestInfo(LOG, req)

	proxyRule := req.URL.Query().Get(QUERYPARAM_PROXYRULE)
	if proxyRule == "" {
		proxyRule = conf.All.HttpService.DefaultProxyRule
	}
	LOG.Debug("using proxy rule: %v", proxyRule)

	rule, found := compiledRules[proxyRule]
	if !found {
		err := fmt.Errorf("proxyRule '%v' is unknown", proxyRule)
		LOG.Error("bad request: %v", err)
		http.Error(respw, fmt.Sprintf("bad request: %v", err), 400)
		return
	}

	fetchUrl := req.URL.Query().Get(QUERYPARAM_METRICS_FETCH_URL)
	if fetchUrl == "" {
		fetchUrl = conf.All.HttpService.DefaultMetricsFetchUrl
	}
	if !strings.HasPrefix(fetchUrl, "http://") {
		fetchUrl = "http://" + fetchUrl
	}
	LOG.Debug("using fetchUrl: %v", fetchUrl)

	// time to do the query!

	LOG.Debug("sending GET request to %v ... (time %v millis)", fetchUrl, time.Since(startTime).Milliseconds())

	resp, err := http.Get(fetchUrl)
	if err != nil || resp.StatusCode >= 300 {
		err := fmt.Errorf("request to '%v' failed! status code was %v, error was: %v", fetchUrl, resp.StatusCode, err)
		LOG.Error("%v", err)
		http.Error(respw, fmt.Sprintf("%v", err), 500)
		return
	}

	LOG.Debug("response arrived from %v, processing... (time %v millis)", fetchUrl, time.Since(startTime).Milliseconds())

	linesProcessed := 0
	responseLines := make([]string, 0)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		linesProcessed++
		if rule.EvaluateLine(line) {
			responseLines = append(responseLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		err := fmt.Errorf("request to '%v' failed! could not read response, error: %v", fetchUrl, err)
		LOG.Error("%v", err)
		http.Error(respw, fmt.Sprintf("%v", err), 500)
		return
	}

	LOG.Debug("response from %v is processed. processed %v lines, matching %v (time %v millis)", fetchUrl, linesProcessed, len(responseLines), time.Since(startTime).Milliseconds())

	respBody := strings.Join(responseLines, "\n")
	fmt.Fprintf(respw, respBody)
}
