package orchestrate

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestOrchestrate(t *testing.T) { TestingT(t) }

type test_context string

func (t test_context) UserId() string {
	return string(t)
}
func (t test_context) UrlParameter(k string) string {
	return ""
}

type OrchestrateTests struct {
	zc zk.ZK
	c  Context
}

var _ = Suite(&OrchestrateTests{})

func (suite *OrchestrateTests) SetUpSuite(c *C) {
	c.Log("Connecting to zk")
	zc, err := zk.Connect([]string{"localhost:2181"}, 5*time.Second)
	c.Assert(err, Equals, nil)
	suite.zc = zc
	suite.c = test_context("test")
	return
}
