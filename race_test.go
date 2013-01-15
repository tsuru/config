// Copyright 2013 config authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build race

package config

import (
	. "launchpad.net/gocheck"
	"sync"
)

func (s *S) TestConfigFunctionsAreThreadSafe(c *C) {
	var wg sync.WaitGroup
	Set("name", "gopher")
	wg.Add(3)
	go func() {
		err := ReadConfigBytes([]byte("name: gopher"))
		if err != nil {
			Get("name")
		}
		wg.Done()
	}()
	go func() {
		Unset("name")
		wg.Done()
	}()
	go func() {
		_, err := GetString("name")
		if err == nil {
			Unset("name")
		} else {
			Set("name", "")
		}
		_, err = GetBool("name")
		if err != nil {
			Set("name", false)
		}
		wg.Done()
	}()
	wg.Wait()
	c.Assert(configs, HasLen, 1)
}
