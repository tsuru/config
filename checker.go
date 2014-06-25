// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
)

// Check a parsed config file.
func Check(config map[interface{}]interface{}) error {
	return CheckProvisioner(config)

}

// Check docker configs
func CheckProvisioner(config map[interface{}]interface{}) error {
	if config["provisioner"] == "docker" {
		return CheckDocker(config)
	}
	return nil
}

func CheckDocker(config map[interface{}]interface{}) error {
	if _, ok := config["docker"]; !ok {
		return errors.New("Config Error: you should configure docker.")
	}
	err := CheckDockerBasicConfig(config["docker"].(map[interface{}]interface{}))
	if err != nil {
		return err
	}
	return nil
}

func CheckDockerBasicConfig(config map[interface{}]interface{}) error {
	basicConfigs := []string{
		"repository-namespace",
		"collection",
		"deploy-cmd",
	}
	for _, key := range basicConfigs {
		if _, ok := config[key]; !ok {
			errorMsg := fmt.Sprintf("Config Error: you should configure %s", key)
			return errors.New(errorMsg)
		}
	}
	return nil
}
