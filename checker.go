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

// Check provisioner configs
func CheckProvisioner() error {
	if configs["provisioner"] == "docker" {
		return CheckDocker()
	}
	return nil
}

// Check Docker configs
func CheckDocker() error {
	if _, err := Get("docker"); err != nil {
		return errors.New("Config Error: you should configure docker.")
	}
	err := CheckDockerBasicConfig()
	if err != nil {
		return err
	}
	err = CheckScheduler()
	if err != nil {
		return err
	}
	return nil
}

// Check default configs to Docker.
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

func CheckScheduler() error {
	if scheduler, err := Get("docker:segregate"); err == nil && scheduler == true {
		if servers, err := Get("docker:servers"); err == nil && servers != nil {
			return fmt.Errorf("Your scheduler is the segregate. Please remove the servers conf in docker.")
		}
		for _, value := range []string{"docker:scheduler:redis-server", "docker:scheduler:redis-prefix"} {
			if _, err := Get(value); err != nil {
				return fmt.Errorf("You should configure %s.", value)
			}
		}
		return nil
	}
	if servers, err := Get("docker:servers"); err != nil || servers == nil {
		return fmt.Errorf("You should configure the docker servers.")
	}
	return nil
}
