package aws

import (
	"encoding/json"
	. "gopkg.in/check.v1"
	"testing"
)

func TestAws(t *testing.T) { TestingT(t) }

type AwsTests struct{}

var _ = Suite(&AwsTests{})

func (vpc *VirtualPrivateCloud) test_get_subnets() []interface{} {
	return []interface{}{
		Subnet{
			Node: Node{
				Id:   "us-west-a",
				Name: "us-west-a",
			},
			CIDR:    "10.31.1.0/20",
			Private: false,
		},
		Subnet{
			Node: Node{
				Id:   "us-west-b",
				Name: "us-west-b",
			},
			CIDR:    "10.31.2.0/20",
			Private: true,
		},
	}
}

func (suite *AwsTests) TestMarshal(c *C) {

	vpc := VirtualPrivateCloud{
		Node: Node{
			Id:   "vpc-1234",
			Name: "prod-vpc",
			Type: "vpc",
		},
	}

	vpc.Node.Children = vpc.test_get_subnets

	m, err := json.Marshal(vpc)
	c.Assert(err, Equals, nil)
	c.Log(string(m))

}
