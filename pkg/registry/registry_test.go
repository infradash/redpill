package registry

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) { TestingT(t) }

type test_context string

func (t test_context) UserId() string {
	return string(t)
}
func (t test_context) UrlParameter(k string) string {
	return ""
}

type RegistryTests struct {
	zc zk.ZK
	c  Context
}

var _ = Suite(&RegistryTests{})

func (suite *RegistryTests) SetUpSuite(c *C) {
	c.Log("Connecting to zk")
	zc, err := zk.Connect([]string{"localhost:2181"}, 5*time.Second)
	c.Assert(err, Equals, nil)
	suite.zc = zc

	zk.CreateOrSet(zc, "/unit-test/registry/test/object1", "object1")
	zk.CreateOrSet(zc, "/unit-test/registry/test/object2", "object2")
	zk.CreateOrSet(zc, "/unit-test/registry/test/object3", "object3")

	suite.c = test_context("test")
	return
}

func (suite *RegistryTests) TestGetEntry(c *C) {
	z := func() zk.ZK { return suite.zc }

	reg := NewService(z)
	c.Log(reg)

	value, rev, err := reg.GetEntry(suite.c, "/unit-test/registry/test/object1")
	c.Assert(err, Equals, nil)
	c.Log("value=", string(value), "rev=", rev)
	c.Assert("object1", DeepEquals, string(value))
}

func (suite *RegistryTests) TestUpdateEntry(c *C) {
	z := func() zk.ZK { return suite.zc }

	reg := NewService(z)
	c.Log(reg)

	_, rev, err := reg.GetEntry(suite.c, "/unit-test/registry/test/object1")
	c.Assert(err, Equals, nil)

	rev2, err := reg.UpdateEntry(suite.c, "/unit-test/registry/test/object1", []byte("new-value"), rev)
	c.Assert(err, Equals, nil)
	c.Assert(rev2, Not(Equals), rev)

	// one more with bad rev
	_, err = reg.UpdateEntry(suite.c, "/unit-test/registry/test/object1", []byte("new-value2"), rev)
	c.Assert(err, Equals, ErrConflict)

	value, rev, err := reg.GetEntry(suite.c, "/unit-test/registry/test/object1")
	c.Assert(value, DeepEquals, []byte("new-value"))
}
