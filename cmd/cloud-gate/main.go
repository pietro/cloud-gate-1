package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Symantec/Dominator/lib/log/serverlogger"
	"github.com/Symantec/cloud-gate/broker"
	"github.com/Symantec/cloud-gate/broker/appconfiguration"
	"github.com/Symantec/cloud-gate/broker/aws"
	"github.com/Symantec/cloud-gate/broker/configuration"
	"github.com/Symantec/cloud-gate/broker/httpd"
	"github.com/Symantec/tricorder/go/tricorder"
)

var (
	configFilename             = flag.String("config", "appConfig.yml", "Configuration filename")
	configurationCheckInterval = flag.Duration("configurationCheckInterval",
		time.Minute*5, "Configuration check interval")
	accountConfigurationUrl = flag.String("accountConfigurationUrl",
		"file:///etc/cloud-gate/accounts.yml", "URL containing account configuration")
)

func main() {
	flag.Parse()
	tricorder.RegisterFlags()
	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "Do not run the cloud-gate server as root")
		os.Exit(1)
	}
	logger := serverlogger.New("")
	configChannel, err := configuration.Watch(*accountConfigurationUrl,
		*configurationCheckInterval, logger)
	if err != nil {
		logger.Fatalf("Cannot watch for configuration: %s\n", err)
	}

	appConfig, err := appconfiguration.LoadVerifyConfigFile(*configFilename)
	if err != nil {
		logger.Fatalf("Cannot load Configuration: %s\n", err)
	}

	logger.Debugf(1, "appconfig=+%v", appConfig)

	brokers := map[string]broker.Broker{
		"aws": aws.New(logger),
	}

	webServer, err := httpd.StartServer(appConfig, brokers, logger)
	if err != nil {
		logger.Fatalf("Unable to create http server: %s\n", err)
	}
	webServer.AddHtmlWriter(logger)
	for config := range configChannel {
		logger.Println("Received new configuration")
		if err := webServer.UpdateConfiguration(config); err != nil {
			logger.Println(err)
		}
		for _, broker := range brokers {
			if err := broker.UpdateConfiguration(config); err != nil {
				logger.Println(err)
			}
		}
	}
}
