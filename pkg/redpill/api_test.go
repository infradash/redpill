package redpill

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestApi(t *testing.T) { TestingT(t) }

type ApiTests struct{}

var _ = Suite(&ApiTests{})

func (suite *ApiTests) TestGetEnvironmentVars(c *C) {
	list := Methods[GetEnvironmentVars].ResponseBody(nil).(EnvList)
	c.Assert(len(list), Equals, 0)

	change := Methods[UpdateEnvironmentVars].RequestBody(nil).(*EnvChange)
	c.Assert(len(change.Update), Equals, 0)
	c.Assert(len(change.Delete), Equals, 0)
}
