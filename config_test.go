// Copyright 2013 config authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	. "launchpad.net/gocheck"
	"os"
	"runtime"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

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
}

func (s *S) TearDownTest(c *C) {
	configs = nil
}

func (s *S) TestConfig(c *C) {
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
`
	err := ReadConfigBytes([]byte(conf))
	c.Assert(err, IsNil)
	c.Assert(configs, DeepEquals, expected)
}

func (s *S) TestConfigFile(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	c.Assert(configs, DeepEquals, expected)
}

func (s *S) TestWriteConfigFile(c *C) {
	Set("database:host", "127.0.0.1")
	Set("database:port", 3306)
	Set("database:user", "root")
	Set("database:password", "s3cr3t")
	Set("database:name", "mydatabase")
	Set("something", "otherthing")
	err := WriteConfigFile("/tmp/config-test.yaml", 0644)
	c.Assert(err, IsNil)
	defer os.Remove("/tmp/config-test.yaml")
	configs = nil
	err = ReadConfigFile("/tmp/config-test.yaml")
	c.Assert(err, IsNil)
	v, err := Get("database:host")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "127.0.0.1")
	v, err = Get("database:port")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, 3306)
	v, err = Get("database:user")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "root")
	v, err = Get("database:password")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "s3cr3t")
	v, err = Get("database:name")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "mydatabase")
	v, err = Get("something")
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "otherthing")
}

func (s *S) TestGetConfig(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := Get("xpto")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "ble")
	value, err = Get("database:host")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "127.0.0.1")
}

func (s *S) TestGetConfigReturnErrorsIfTheKeyIsNotFound(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := Get("xpta")
	c.Assert(value, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `key "xpta" not found`)
	value, err = Get("database:hhh")
	c.Assert(value, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `key "database:hhh" not found`)
}

func (s *S) TestGetString(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := GetString("xpto")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "ble")
	value, err = GetString("database:host")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "127.0.0.1")
}

func (s *S) TestGetStringShouldReturnErrorIfTheKeyDoesNotRepresentAString(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := GetString("database:port")
	c.Assert(value, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `value for the key "database:port" is not a string`)
}

func (s *S) TestGetStringShouldReturnErrorIfTheKeyDoesNotExist(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := GetString("xpta")
	c.Assert(value, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `key "xpta" not found`)
}

func (s *S) TestGetBool(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := GetBool("istrue")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, false)
}

func (s *S) TestGetBoolWithNonBoolConfValue(c *C) {
	configFile := "testdata/config.yml"
	err := ReadConfigFile(configFile)
	c.Assert(err, IsNil)
	value, err := GetBool("fakebool")
	c.Assert(value, Equals, false)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `value for the key "fakebool" is not a boolean`)
}

func (s *S) TestGetList(c *C) {
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
	c.Assert(err, IsNil)
	Set("dynamic", []string{"Mary", "Petter"})
	for _, t := range tests {
		values, err := GetList(t.key)
		c.Check(err, DeepEquals, t.err)
		c.Check(values, DeepEquals, t.expected)
	}
}

func (s *S) TestSet(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	Set("xpto", "bla")
	value, err := GetString("xpto")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "bla")
}

func (s *S) TestSetChildren(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	Set("database:host", "database.com")
	value, err := GetString("database:host")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "database.com")
}

func (s *S) TestSetChildrenDoesNotImpactOtherChild(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	Set("database:host", "database.com")
	value, err := Get("database:port")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, 8080)
}

func (s *S) TestSetMap(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	Set("database", map[interface{}]interface{}{"host": "database.com", "port": 3306})
	host, err := GetString("database:host")
	c.Assert(err, IsNil)
	c.Assert(host, Equals, "database.com")
	port, err := Get("database:port")
	c.Assert(err, IsNil)
	c.Assert(port, Equals, 3306)
}

func (s *S) TestUnset(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	err = Unset("xpto")
	c.Assert(err, IsNil)
	_, err = Get("xpto")
	c.Assert(err, NotNil)
}

func (s *S) TestUnsetChildren(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	err = Unset("database:host")
	c.Assert(err, IsNil)
	_, err = Get("database:host")
	c.Assert(err, NotNil)
}

func (s *S) TestUnsetWithUndefinedKey(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	err = Unset("database:hoster")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `Key "database:hoster" not found`)
}

func (s *S) TestUnsetMap(c *C) {
	err := ReadConfigFile("testdata/config.yml")
	c.Assert(err, IsNil)
	err = Unset("database")
	c.Assert(err, IsNil)
	_, err = Get("database:host")
	c.Assert(err, NotNil)
	_, err = Get("database:port")
	c.Assert(err, NotNil)
}

func (s *S) TestMergeMaps(c *C) {
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
	c.Assert(mergeMaps(m1, m2), DeepEquals, expected)
}

func (s *S) TestMergeMapsMultipleProcs(c *C) {
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
	c.Assert(mergeMaps(m1, m2), DeepEquals, expected)
}

func (s *S) TestMergeMapsWithDiffingMaps(c *C) {
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
	c.Assert(mergeMaps(m1, m2), DeepEquals, expected)
}
