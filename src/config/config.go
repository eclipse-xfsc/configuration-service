package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               int
	ConfigmapName      string
	ConfigmapNamespace string
}

// Load reads configuration from environment variables
func Load() (Config, error) {
	port, found := os.LookupEnv("PORT")
	if !found {
		return Config{}, fmt.Errorf("environment variable \"PORT\" not found")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return Config{}, err
	}

	configmapName, found := os.LookupEnv("CONFIGMAP_NAME")
	if !found {
		return Config{}, fmt.Errorf("environment variable \"CONFIGMAP_NAME\" not found")
	}

	configmapNamespace, found := os.LookupEnv("CONFIGMAP_NAMESPACE")
	if !found {
		return Config{}, fmt.Errorf("environment variable \"CONFIGMAP_NAMESPACE\" not found")
	}

	return Config{
		Port:               portInt,
		ConfigmapName:      configmapName,
		ConfigmapNamespace: configmapNamespace,
	}, nil
}
