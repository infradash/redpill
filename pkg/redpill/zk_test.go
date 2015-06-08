package redpill

import (
	"github.com/qorio/maestro/pkg/zk"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestZk(t *testing.T) { TestingT(t) }

type ZkTests struct{}

var _ = Suite(&ZkTests{})

func (suite *ZkTests) TestGetEnvironments(c *C) {
	z := Zk{
		Hosts:   "localhost:2181",
		Timeout: 5 * time.Second,
	}

	err := z.Connect()
	c.Assert(err, Equals, nil)

	_, _, err = z.GetEnv("xxxops-test.blinker.com", "blinker", "develop")
	c.Assert(err, Equals, zk.ErrNotExist)

	list, rev, err := z.GetEnv("ops-test.blinker.com", "blinker", "develop")
	c.Assert(err, Equals, nil)
	c.Log("revision", rev)
	c.Assert(rev, Not(Equals), 0)
	c.Log("list", list)
}
