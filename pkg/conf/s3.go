package conf

import (
	"errors"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"path"
)

var (
	ErrBadConfig = errors.New("bad-config")
	ErrBadRegion = errors.New("bad-region")
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

	bucket *s3.Bucket
	auth   *aws.Auth
	s3     *s3.S3
}

func (this *S3Bucket) Init() error {
	_, region_ok := aws_regions[this.Region]
	switch {
	case this.Region == "", this.Bucket == "", this.AccessKey == "", this.AccessToken == "":
		return ErrBadConfig
	case !region_ok:
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

// ConfStorage interface
func (this *S3Bucket) Save(domainClass, service, name string, content []byte) error {
	key := path.Join(domainClass, service, name)
	return this.bucket.Put(key, content, "text/plain", s3.Private)
}

// ConfStorage interface
func (this *S3Bucket) Get(domainClass, service, name string) ([]byte, error) {
	key := path.Join(domainClass, service, name)
	return this.bucket.Get(key)
}

// ConfStorage interface
func (this *S3Bucket) Delete(domainClass, service, name string) error {
	key := path.Join(domainClass, service, name)
	return this.bucket.Del(key)
}

// ConfStorage interface
func (this *S3Bucket) SaveVersion(domainClass, domainInstance, service, version, name string, content []byte) error {
	key := path.Join(domainClass, domainInstance, service, version, name)
	return this.bucket.Put(key, content, "text/plain", s3.Private)
}

// ConfStorage interface
func (this *S3Bucket) GetVersion(domainClass, domainInstance, service, version, name string) ([]byte, error) {
	key := path.Join(domainClass, domainInstance, service, version, name)
	return this.bucket.Get(key)
}

// ConfStorage interface
func (this *S3Bucket) DeleteVersion(domainClass, domainInstance, service, version, name string) error {
	key := path.Join(domainClass, domainInstance, service, version, name)
	return this.bucket.Del(key)

}
