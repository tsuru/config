// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"launchpad.net/gocheck"
)

type CheckerSuite struct{}

var _ = gocheck.Suite(&CheckerSuite{})

func (s *CheckerSuite) TestCheckExecuteCheckerFunc(c *gocheck.C) {
	err := Check([]Checker{
		func() error { return errors.New("Fake checker error") },
	})
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.DeepEquals, "Fake checker error")
}
