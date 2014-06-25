// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"launchpad.net/gocheck"
)

type CheckerSuite struct{}

var _ = gocheck.Suite(&CheckerSuite{})

var configFixture = `
listen: 0.0.0.0:8080
host: http://127.0.0.1:8080
debug: true
admin-team: admin

database:
  url: 127.0.0.1:3435
  name: tsuru

git:
  unit-repo: /home/application/current
  api-server: http://127.0.0.1:8000
  rw-host: 127.0.0.1
  ro-host: 127.0.0.1

auth:
  user-registration: true
  scheme: native

provisioner: docker
hipache:
  domain: tsuru-sample.com
queue: redis
redis-queue:
  host: localhost
  port: 6379
docker:
  collection: docker_containers
  repository-namespace: tsuru
  router: hipache
  deploy-cmd: /var/lib/tsuru/deploy
  ssh-agent-port: 4545
  segregate: true
  scheduler:
    redis-server: 127.0.0.1:6379
    redis-prefix: docker-cluster
  run-cmd:
    bin: /var/lib/tsuru/start
    port: 8888
  ssh:
    add-key-cmd: /var/lib/tsuru/add-key
    public-key: /var/lib/tsuru/.ssh/id_rsa.pub
    user: ubuntu
`

func (s *CheckerSuite) SetUpTest(c *gocheck.C) {
	err := ReadConfigBytes([]byte(configFixture))
	c.Assert(err, gocheck.IsNil)
}

func (s *CheckerSuite) TearDownTest(c *gocheck.C) {
	configs = nil
}

func (s *CheckerSuite) TestCheckConfig(c *gocheck.C) {
	err := Check(configs)
	c.Assert(err, gocheck.IsNil)
}

func (s *CheckerSuite) TestCheckDockerJustCheckIfProvisionerIsDocker(c *gocheck.C) {
	configs["provisioner"] = "test"
	err := CheckProvisioner(configs)
	c.Assert(err, gocheck.IsNil)
}

func (s *CheckerSuite) TestCheckDockerIsNotConfigured(c *gocheck.C) {
	delete(configs, "docker")
	err := CheckDocker(configs)
	c.Assert(err, gocheck.NotNil)
}

func (s *CheckerSuite) TestCheckDockerBasicConfig(c *gocheck.C) {
	err := CheckDockerBasicConfig(configs["docker"].(map[interface{}]interface{}))
	c.Assert(err, gocheck.IsNil)
}

func (s *CheckerSuite) TestCheckDockerBasicConfigError(c *gocheck.C) {
	delete(configs["docker"].(map[interface{}]interface{}), "collection")
	err := CheckDockerBasicConfig(configs["docker"].(map[interface{}]interface{}))
	c.Assert(err, gocheck.NotNil)
}
