// Copyright 2015 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"

	"gopkg.in/check.v1"
)

type CheckerSuite struct{}

var _ = check.Suite(&CheckerSuite{})

func (s *CheckerSuite) TestCheckExecuteCheckerFunc(c *check.C) {
	err := Check([]Checker{
		func() error { return errors.New("Fake checker error") },
	})
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.DeepEquals, "Fake checker error")
}
