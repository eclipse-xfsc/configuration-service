package main

import (
	"fmt"
	"os"
	"strconv"
)

type config struct {
	port int
	configmapName string
	configmapNamespace string
}

func getConfig() (config, error) {
	port, found := os.LookupEnv("PORT")
    if !found {
		err := fmt.Errorf("Environemnt variable \"PORT\" not found")
		return config{}, err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return config{}, err
	}

	configmapName, found := os.LookupEnv("CONFIGMAP_NAME")
    if !found {
		err := fmt.Errorf("Environemnt variable \"CONFIGMAP_NAME\" not found")
		return config{}, err
	}

	configmapNamespace, found := os.LookupEnv("CONFIGMAP_NAMESPACE")
    if !found {
		err := fmt.Errorf("Environemnt variable \"CONFIGMAP_NAMESPACE\" not found")
		return config{}, err
	}


	config := config{
		port: portInt,
		configmapName: configmapName,
		configmapNamespace: configmapNamespace,
	}
	
	return config, nil
}