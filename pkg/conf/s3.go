package conf

import (
	"errors"
	"github.com/golang/glog"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"path"
	"strings"
)

var (
	ErrBadConfig = errors.New("bad-config")
	ErrBadRegion = errors.New("bad-region")
)

var (
	RegionUSGovWest    = aws.USGovWest
	RegionUSEast1      = aws.USEast
	RegionUSWest1      = aws.USWest
	RegionUSWest2      = aws.USWest2
	RegionEUCentral1   = aws.EUCentral
	RegionEUWest1      = aws.EUWest
	RegionAPSouthEast1 = aws.APSoutheast
	RegionAPSouthEast2 = aws.APSoutheast2
)

var (
	// https://github.com/mitchellh/goamz/blob/master/aws/aws.go
	aws_regions = map[string]aws.Region{
		"us-gov-west-1":  aws.USGovWest,
		"us-east-1":      aws.USEast,
		"us-west-1":      aws.USWest,
		"us-west-2":      aws.USWest2,
		"eu-west-1":      aws.EUWest,
		"eu-central-1":   aws.EUCentral,
		"ap-southeast-1": aws.APSoutheast,
		"ap-southeast-2": aws.APSoutheast2,
		"sa-east-1":      aws.SAEast,
	}
)

type S3Bucket struct {
	Region      string `json:"aws_region"`
	Bucket      string `json:"bucket"`
	AccessKey   string `json:"access_key"`
	AccessToken string `json:"access_token"`
	BucketRoot  string `json:"bucket_root"`

	bucket *s3.Bucket
	auth   *aws.Auth
	s3     *s3.S3
	zk     zk.ZK
}

func (this *S3Bucket) IsRequested() bool {
	return len(this.Bucket) > 0
}

func (this *S3Bucket) Init(pool func() zk.ZK) error {
	this.zk = pool()

	if !this.get_envs() {
		glog.Infoln("BadEnvs")
		return ErrBadConfig
	}

	if _, region_ok := aws_regions[this.Region]; !region_ok {
		glog.Infoln("BadRegion")
		return ErrBadRegion
	}

	auth, err := aws.GetAuth(this.AccessKey, this.AccessToken)
	if err != nil {
		return err
	}
	this.auth = &auth
	this.s3 = s3.New(auth, aws_regions[this.Region])
	this.bucket = this.s3.Bucket(this.Bucket)
	return nil
}

func (this *S3Bucket) get_envs() bool {
	return update_var(this.zk, &this.Region) &&
		update_var(this.zk, &this.AccessKey) &&
		update_var(this.zk, &this.AccessToken)
}

const (
	prefix = "zk://"
)

func update_var(zc zk.ZK, v *string) bool {
	switch strings.Index(*v, prefix) {
	case 0:
		key := *v
		if value := zk.GetString(zc, registry.Path(key[len(prefix):])); value != nil {
			*v = *value
			return true
		}
		return false
	case -1:
		return true
	}
	return false
}

// ConfStorage interface
func (this *S3Bucket) Save(domainClass, service, name string, content []byte) error {
	key := path.Join(this.BucketRoot, domainClass, "_base", service, name)
	return this.bucket.Put(key, content, "text/plain", s3.Private)
}

// ConfStorage interface
func (this *S3Bucket) Get(domainClass, service, name string) ([]byte, error) {
	key := path.Join(this.BucketRoot, domainClass, "_base", service, name)
	return this.bucket.Get(key)
}

// ConfStorage interface
func (this *S3Bucket) Delete(domainClass, service, name string) error {
	key := path.Join(this.BucketRoot, domainClass, "_base", service, name)
	return this.bucket.Del(key)
}

// ConfStorage interface
func (this *S3Bucket) SaveVersion(domainClass, domainInstance, service, version, name string, content []byte) error {
	key := path.Join(this.BucketRoot, domainClass, domainInstance, service, version, name)
	return this.bucket.Put(key, content, "text/plain", s3.Private)
}

// ConfStorage interface
func (this *S3Bucket) GetVersion(domainClass, domainInstance, service, version, name string) ([]byte, error) {
	key := path.Join(this.BucketRoot, domainClass, domainInstance, service, version, name)
	return this.bucket.Get(key)
}

// ConfStorage interface
func (this *S3Bucket) DeleteVersion(domainClass, domainInstance, service, version, name string) error {
	key := path.Join(this.BucketRoot, domainClass, domainInstance, service, version, name)
	return this.bucket.Del(key)

}
