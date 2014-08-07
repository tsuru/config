// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"launchpad.net/gocheck"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct{}

var _ = gocheck.Suite(&S{})

var expected = map[interface{}]interface{}{
	"database": map[interface{}]interface{}{
		"host": "127.0.0.1",
		"port": 8080,
	},
	"auth": map[interface{}]interface{}{
		"salt": "xpto",
		"key":  "sometoken1234",
	},
	"xpto":           "ble",
	"istrue":         false,
	"fakebool":       "foo",
	"names":          []interface{}{"Mary", "John", "Anthony", "Gopher"},
	"multiple-types": []interface{}{"Mary", 50, 5.3, true},
	"negative":       -10,
}

func (s *S) TearDownTest(c *gocheck.C) {
	configs = nil
}

func (s *S) TestConfig(c *gocheck.C) {
	conf := `
database:
  host: 127.0.0.1
  port: 8080
auth:
  salt: xpto
  key: sometoken1234
xpto: ble
istrue: false
fakebool: foo
names:
  - Mary
  - John
  - Anthony
  - Gopher
multiple-types:
  - Mary
  - 50
  - 5.3
  - true
negative: -10
`
	err := ReadConfigBytes([]byte(conf))
	c.Assert(err, gocheck.IsNil)
	c.Assert(configs, gocheck.DeepEquals, expected)
}

func (s *S) TestConfigFile(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	c.Assert(configs, gocheck.DeepEquals, expected)
}

func (s *S) TestConfigFileUnknownFile(c *gocheck.C) {
	err := ReadConfigFile("/some/unknwon/file/path")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestWatchConfigFile(c *gocheck.C) {
	err := exec.Command("cp", "testdata/config.yml", "/tmp/config-test.yml").Run()
	c.Assert(err, gocheck.IsNil)
	err = ReadAndWatchConfigFile("/tmp/config-test.yml")
	c.Assert(err, gocheck.IsNil)
	mut.Lock()
	c.Check(configs, gocheck.DeepEquals, expected)
	mut.Unlock()
	err = exec.Command("cp", "testdata/config2.yml", "/tmp/config-test.yml").Run()
	c.Assert(err, gocheck.IsNil)
	time.Sleep(1e9)
	expectedAuth := map[interface{}]interface{}{
		"salt": "xpta",
		"key":  "sometoken1234",
	}
	mut.Lock()
	c.Check(configs["auth"], gocheck.DeepEquals, expectedAuth)
	mut.Unlock()
}

func (s *S) TestWatchConfigFileUnknownFile(c *gocheck.C) {
	err := ReadAndWatchConfigFile("/some/unknwon/file/path")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestWriteConfigFile(c *gocheck.C) {
	Set("database:host", "127.0.0.1")
	Set("database:port", 3306)
	Set("database:user", "root")
	Set("database:password", "s3cr3t")
	Set("database:name", "mydatabase")
	Set("something", "otherthing")
	err := WriteConfigFile("/tmp/config-test.yaml", 0644)
	c.Assert(err, gocheck.IsNil)
	defer os.Remove("/tmp/config-test.yaml")
	configs = nil
	err = ReadConfigFile("/tmp/config-test.yaml")
	c.Assert(err, gocheck.IsNil)
	v, err := Get("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, "127.0.0.1")
	v, err = Get("database:port")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, 3306)
	v, err = Get("database:user")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, "root")
	v, err = Get("database:password")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, "s3cr3t")
	v, err = Get("database:name")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, "mydatabase")
	v, err = Get("something")
	c.Assert(err, gocheck.IsNil)
	c.Assert(v, gocheck.Equals, "otherthing")
}

func (s *S) TestGetConfig(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := Get("xpto")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "ble")
	value, err = Get("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "127.0.0.1")
}

func (s *S) TestGetConfigReturnErrorsIfTheKeyIsNotFound(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := Get("xpta")
	c.Assert(value, gocheck.IsNil)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `key "xpta" not found`)
	value, err = Get("database:hhh")
	c.Assert(value, gocheck.IsNil)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `key "database:hhh" not found`)
}

func (s *S) TestGetConfigExpandVars(c *gocheck.C) {
	configFile := "testdata/config3.yml"
	err := os.Setenv("DBHOST", "6.6.6.6")
	defer os.Setenv("DBHOST", "")
	c.Assert(err, gocheck.IsNil)
	err = ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := Get("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "6.6.6.6")
}

func (s *S) TestGetStringExpandVars(c *gocheck.C) {
	configFile := "testdata/config3.yml"
	err := os.Setenv("DBHOST", "6.6.6.6")
	defer os.Setenv("DBHOST", "")
	c.Assert(err, gocheck.IsNil)
	err = ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetString("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "6.6.6.6")
}

func (s *S) TestGetString(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetString("xpto")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "ble")
	value, err = GetString("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "127.0.0.1")
}

func (s *S) TestGetInt(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetInt("database:port")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, 8080)
	value, err = GetInt("xpto")
	c.Assert(err, gocheck.NotNil)
	value, err = GetInt("something-unknown")
	c.Assert(err, gocheck.NotNil)
	c.Assert(value, gocheck.Equals, 0)
}

func (s *S) TestGetUint(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetUint("database:port")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, uint(8080))
	_, err = GetUint("negative")
	c.Assert(err, gocheck.NotNil)
	_, err = GetUint("auth:salt")
	c.Assert(err, gocheck.NotNil)
	_, err = GetUint("Unknown")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestGetStringShouldReturnErrorIfTheKeyDoesNotRepresentAString(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetString("database:port")
	c.Assert(value, gocheck.Equals, "")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, `value for the key "database:port" is not a string`)
}

func (s *S) TestGetStringShouldReturnErrorIfTheKeyDoesNotExist(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetString("xpta")
	c.Assert(value, gocheck.Equals, "")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, `key "xpta" not found`)
}

func (s *S) TestGetDuration(c *gocheck.C) {
	configFile := "testdata/config2.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetDuration("interval")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(1e9))
	value, err = GetDuration("superinterval")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(1e9))
	value, err = GetDuration("wait")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(1e6))
	value, err = GetDuration("one_year")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(365*24*time.Hour))
	value, err = GetDuration("nano")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(1))
	value, err = GetDuration("human-interval")
	c.Check(err, gocheck.IsNil)
	c.Check(value, gocheck.Equals, time.Duration(10e9))
}

func (s *S) TestGetDurationUnknown(c *gocheck.C) {
	configFile := "testdata/config2.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetDuration("intervalll")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, `key "intervalll" not found`)
	c.Assert(value, gocheck.Equals, time.Duration(0))
}

func (s *S) TestGetDurationInvalid(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetDuration("auth:key")
	c.Assert(value, gocheck.Equals, time.Duration(0))
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, `value for the key "auth:key" is not a duration`)
}

func (s *S) TestGetBool(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetBool("istrue")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, false)
}

func (s *S) TestGetBoolWithNonBoolConfValue(c *gocheck.C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, gocheck.IsNil)
	value, err := GetBool("fakebool")
	c.Assert(value, gocheck.Equals, false)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, `value for the key "fakebool" is not a boolean`)
}

func (s *S) TestGetBoolUndeclaredValue(c *gocheck.C) {
	value, err := GetBool("something-unknown")
	c.Assert(value, gocheck.Equals, false)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `key "something-unknown" not found`)
}

func (s *S) TestGetList(c *gocheck.C) {
	var tests = []struct {
		key      string
		expected []string
		err      error
	}{
		{
			key:      "names",
			expected: []string{"Mary", "John", "Anthony", "Gopher"},
			err:      nil,
		},
		{
			key:      "multiple-types",
			expected: []string{"Mary", "50", "5.3", "true"},
			err:      nil,
		},
		{
			key:      "fakebool",
			expected: nil,
			err:      &invalidValue{"fakebool", "list"},
		},
		{
			key:      "dynamic",
			expected: []string{"Mary", "Petter"},
			err:      nil,
		},
	}
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	Set("dynamic", []string{"Mary", "Petter"})
	for _, t := range tests {
		values, err := GetList(t.key)
		c.Check(err, gocheck.DeepEquals, t.err)
		c.Check(values, gocheck.DeepEquals, t.expected)
	}
}

func (s *S) TestGetListUndeclaredValue(c *gocheck.C) {
	value, err := GetList("something-unknown")
	c.Assert(value, gocheck.IsNil)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `key "something-unknown" not found`)
}

func (s *S) TestGetListWithStringers(c *gocheck.C) {
	err := errors.New("failure")
	Set("what", []interface{}{err})
	value, err := GetList("what")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.DeepEquals, []string{"failure"})
}

func (s *S) TestSet(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	Set("xpto", "bla")
	value, err := GetString("xpto")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "bla")
}

func (s *S) TestSetChildren(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	Set("database:host", "database.com")
	value, err := GetString("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "database.com")
}

func (s *S) TestSetChildrenDoesNotImpactOtherChild(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	Set("database:host", "database.com")
	value, err := Get("database:port")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, 8080)
}

func (s *S) TestSetMap(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	Set("database", map[interface{}]interface{}{"host": "database.com", "port": 3306})
	host, err := GetString("database:host")
	c.Assert(err, gocheck.IsNil)
	c.Assert(host, gocheck.Equals, "database.com")
	port, err := Get("database:port")
	c.Assert(err, gocheck.IsNil)
	c.Assert(port, gocheck.Equals, 3306)
}

func (s *S) TestUnset(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	err = Unset("xpto")
	c.Assert(err, gocheck.IsNil)
	_, err = Get("xpto")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestUnsetChildren(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	err = Unset("database:host")
	c.Assert(err, gocheck.IsNil)
	_, err = Get("database:host")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestUnsetWithUndefinedKey(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	err = Unset("database:hoster")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `Key "database:hoster" not found`)
}

func (s *S) TestUnsetMap(c *gocheck.C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, gocheck.IsNil)
	err = Unset("database")
	c.Assert(err, gocheck.IsNil)
	_, err = Get("database:host")
	c.Assert(err, gocheck.NotNil)
	_, err = Get("database:port")
	c.Assert(err, gocheck.NotNil)
}

func (s *S) TestMergeMaps(c *gocheck.C) {
	m1 := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "localhost",
			"port": 3306,
		},
	}
	m2 := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "remotehost",
		},
		"memcached": []string{"mymemcached"},
	}
	expected := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "remotehost",
			"port": 3306,
		},
		"memcached": []string{"mymemcached"},
	}
	c.Assert(mergeMaps(m1, m2), gocheck.DeepEquals, expected)
}

func (s *S) TestMergeMapsMultipleProcs(c *gocheck.C) {
	old := runtime.GOMAXPROCS(16)
	defer runtime.GOMAXPROCS(old)
	m1 := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "localhost",
			"port": 3306,
		},
	}
	m2 := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "remotehost",
		},
		"memcached": []string{"mymemcached"},
	}
	expected := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "remotehost",
			"port": 3306,
		},
		"memcached": []string{"mymemcached"},
	}
	c.Assert(mergeMaps(m1, m2), gocheck.DeepEquals, expected)
}

func (s *S) TestMergeMapsWithDiffingMaps(c *gocheck.C) {
	m1 := map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host": "localhost",
			"port": 3306,
		},
	}
	m2 := map[interface{}]interface{}{
		"auth": map[interface{}]interface{}{
			"user":     "root",
			"password": "123",
		},
	}
	expected := map[interface{}]interface{}{
		"auth": map[interface{}]interface{}{
			"user":     "root",
			"password": "123",
		},
		"database": map[interface{}]interface{}{
			"host": "localhost",
			"port": 3306,
		},
	}
	c.Assert(mergeMaps(m1, m2), gocheck.DeepEquals, expected)
}
