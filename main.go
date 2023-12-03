package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ktlogging "github.com/keytiles/lib-logging-golang"
	"github.com/keytiles/prometheus-metrics-filter/pkg/conf"
	http_metrics_api "github.com/keytiles/prometheus-metrics-filter/pkg/http"
	"github.com/keytiles/prometheus-metrics-filter/pkg/rules"
)

const (
	// command line argument names
	arg_ConfigPath    = "cfg"
	arg_LogConfigPath = "logCfg"

	// used environment variable names
	envvar_ConfigPath    = "LOGSERVICE_CFG_PATH"
	envvar_LogConfigPath = "LOGSERVICE_LOG_CFG_PATH"
)

var (
	argCfgPath    = flag.String(arg_ConfigPath, "", "path to service config file")
	argLogCfgPath = flag.String(arg_LogConfigPath, "", "path to service logging config file")
)

func getLoggingConfigPath() string {
	logCfgPath := *argLogCfgPath
	if logCfgPath == "" {
		logCfgPath = os.Getenv(envvar_LogConfigPath)
	}
	if logCfgPath == "" {
		// let's use default
		logCfgPath = "/conf/log-config.yaml"
	}
	return logCfgPath
}

func getConfigPath() string {
	cfgPath := *argCfgPath
	if cfgPath == "" {
		cfgPath = os.Getenv(envvar_ConfigPath)
	}
	if cfgPath == "" {
		cfgPath = "/conf/config.yaml"
	}
	return cfgPath
}

var (
	httpServer *http.Server
)

func startHttpServer() {

	LOGGER := ktlogging.GetLogger("main.startHttpServer")

	httpAddr := conf.All.HttpService.Address + ":" + conf.All.HttpService.Port

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/metrics", http_metrics_api.Api_ExecuteMetricsRequest)

	httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: serveMux,
	}
	// we start it in the background
	go func() {
		LOGGER.Info("firing up httpServer on %v", httpAddr)

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			LOGGER.Error("Oops! Failed to start httpServer due to error: %v", err)
			panic("Critical error occured - check logs! Exiting service")
		}
	}()
}

func getCompileRulesConfig() (compiledRules map[string]rules.ILineEvaluationRule) {
	LOGGER := ktlogging.GetLogger("main.compileRulesConfig")

	compiledRules = make(map[string]rules.ILineEvaluationRule, len(conf.All.ProxyRules))
	for key, proxyRule := range conf.All.ProxyRules {
		compiledRule, failure := rules.NewIncludeRemoveRule(proxyRule)
		if failure != nil {
			LOGGER.Error("problem in config at rule set '%v': %v", key, failure)
			panic("Critical error occured - check logs! Exiting service")
		}
		compiledRules[key] = compiledRule
	}

	return
}

func main() {
	// Parse command-line arguments
	flag.Parse()

	logCfgPath := getLoggingConfigPath()
	logCfgErr := ktlogging.InitFromConfig(logCfgPath)
	if logCfgErr != nil {
		panic(fmt.Sprintf("failed to configure logging! error was: %v", logCfgErr))
	}

	// let's create our package logger
	_LOGGER := ktlogging.GetLogger("main")

	_LOGGER.Info("logging is initialized from config '%v'", logCfgPath)

	cfgPath := getConfigPath()
	_LOGGER.Info("reading up service config '%v' ...", cfgPath)
	err := conf.InitConfig(cfgPath)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config! error: %v\n", err))
	}
	_LOGGER.Info("done!")
	_LOGGER.Info("service configuration is: %+v", conf.All)

	// ok let's rock!

	// we need the rules from the config
	compiledRules := getCompileRulesConfig()
	// pass it to the http api
	http_metrics_api.Init(compiledRules)

	// now we are ready to fire up our own http server
	startHttpServer()

	// let's wait now the exit signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	_LOGGER.Info("Now waiting for kill signal...")
	<-done // Will block here until user hits ctrl+c

	_LOGGER.Info("kill signal arrived - exiting...")

	// shutdown http - right away (so not gracefully with .Shutdown())
	httpServer.Close()

}
