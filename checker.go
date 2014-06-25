// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
)

// Check a parsed config file.
func Check() error {
	return CheckProvisioner()

}

// Check docker configs
func CheckProvisioner() error {
	if configs["provisioner"] == "docker" {
		return CheckDocker()
	}
	return nil
}

func CheckDocker() error {
	if _, err := Get("docker"); err != nil {
		return errors.New("Config Error: you should configure docker.")
	}
	err := CheckDockerBasicConfig()
	if err != nil {
		return err
	}
	return nil
}

func CheckDockerBasicConfig() error {
	basicConfigs := []string{
		"docker:repository-namespace",
		"docker:collection",
		"docker:deploy-cmd",
		"docker:ssh-agent-port",
		"docker:ssh",
		"docker:ssh:add-key-cmd",
		"docker:ssh:public-key",
		"docker:ssh:user",
		"docker:run-cmd:bin",
		"docker:run-cmd:port",
	}
	for _, key := range basicConfigs {
		if _, err := Get(key); err != nil {
			return fmt.Errorf("Config Error: you should configure %s", key)
		}
	}
	return nil
}
