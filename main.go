package main

import (
	"os"

	"github.com/eclipse-xfsc/configuration-service/src/config"
	"github.com/eclipse-xfsc/configuration-service/src/kubernetes"
	"github.com/eclipse-xfsc/configuration-service/src/routes"
	"github.com/eclipse-xfsc/configuration-service/src/telemetry"
)

func init() {
	// Set up test injection functions
	routes.ConfigLoadFn = func() (string, string, error) {
		cfg, err := config.Load()
		return cfg.ConfigmapName, cfg.ConfigmapNamespace, err
	}
	routes.ConfigMapFn = kubernetes.GetConfigmap
}

func main() {
	// Initialize logger
	telemetry.InitializeLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		telemetry.Logger.Error(err)
		os.Exit(1)
	}

	// Start REST API server
	routes.Start(cfg.Port)
}
