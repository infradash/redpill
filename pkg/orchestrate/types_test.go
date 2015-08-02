package orchestrate

import (
	"encoding/json"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestTypes(t *testing.T) { TestingT(t) }

type TypesTests struct {
	zc zk.ZK
	c  Context
}

var _ = Suite(&TypesTests{})

func (suite *TypesTests) SetUpSuite(c *C) {
	c.Log("Connecting to zk")
	zc, err := zk.Connect([]string{"localhost:2181"}, 5*time.Second)
	c.Assert(err, Equals, nil)
	suite.zc = zc
	return
}

func (suite *TypesTests) TestUnmarshalModel(c *C) {

	input := `
{
    "name": "aws-ec2-describe-instances",
    "friendly_name" : "AWS EC2 CLI describe instances",
    "description" : "Describes current instances in AWS EC2",

    "cmd": {
	"path" : "aws",
	"args" : [ "ec2", "describe-intsances", "{{.Context.InstanceId}}", "{{.Context.Region}}", "{{.Env.EXEC_TS}}" ]
    },

    "info": "/{{.Domain}}/task/aws-ec2-describe-instances/{{.Task.Id}}",
    "status": "mqtt://iot.eclipse.org:1883/test.com/task/aws-ec2-describe-instances/{{.Task.Id}}",
    "stdout": "mqtt://iot.eclipse.org:1883/test.com/task/aws-ec2-describe-instances/{{.Task.Id}}/stdout",
    "stderr": "mqtt://iot.eclipse.org:1883/test.com/task/aws-ec2-describe-instances/{{.Task.Id}}/stdin",
    "print_pre" : "Starting up AWS EC2 Describe Instances {{.Env.EXEC_TS}}",
    "print_post" : "Finished describing instances",
    "print_err" : "Error describing instances",


    "default_context" : {
	"InstanceId" : "i-1289ac2d",
	"Region" : "us-west"
    },

    "tail": [
        {
            "path": "/Users/david/go/src/github.com/infradash/dash/example/test2.log",
	    "stderr": true,
	    "topic": "mqtt://iot.eclipse.org:1883/{{.Domain}}/{{.Service}}/{{.Host}}/test.log"
        }
    ],

    "docker" : {
	"name": "{{.Service}}-{{.Version}}-{{.Build}}-{{.Sequence}}",
	"Image": "infradash/aws-cli:106",
	"Env" : [ "DASH_ZK_HOSTS={{zk_hosts}}",
                  "DASH_DOMAIN={{.Domain}}", "DASH_SERVICE=task",
                  "DASH_CONFIG_URL={{config_url}}",
                  "DASH_CMD=''",
                  "DASH_OPTIONS='--daemon=fase --logtostderr'" ],
	"Cmd" : [ ],
	"host_config" : { "PublishAllPorts" : true }
    }
}
`
	m := Model{}

	err := json.Unmarshal([]byte(input), &m)
	c.Assert(err, Equals, nil)
	c.Log(m.Task)

	c.Assert(string(m.Name), DeepEquals, "aws-ec2-describe-instances")
	c.Assert(string(m.Task.Name), DeepEquals, "aws-ec2-describe-instances")

	c.Assert(m.Docker, Not(Equals), nil)
	c.Assert(m.Docker.Image, Equals, "infradash/aws-cli:106")
	c.Assert(m.DefaultContext["InstanceId"], DeepEquals, "i-1289ac2d")

	c.Assert(m.TailFiles[0].Stderr, Equals, true)

	c.Log(m.Task.Info)

	buff, err := json.Marshal(m)
	c.Assert(err, Equals, nil)
	c.Log(string(buff))
}
