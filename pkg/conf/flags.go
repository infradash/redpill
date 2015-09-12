package conf

import (
	"flag"
	. "github.com/qorio/omni/runtime"
)

const (
	EnvS3Region      = "REDPILL_S3_REGION"
	EnvS3Bucket      = "REDPILL_S3_BUCKET"
	EnvS3BucketRoot  = "REDPILL_S3_BUCKET_ROOT"
	EnvS3AccessKey   = "REDPILL_S3_ACCESS_KEY"
	EnvS3AccessToken = "REDPILL_S3_ACCESS_TOKEN"
)

func (this *S3Bucket) BindFlags() {
	flag.StringVar(&this.Region, "s3_region", EnvString(EnvS3Region, "us-west-2"), "S3 Region")
	flag.StringVar(&this.Bucket, "s3_bucket", EnvString(EnvS3Bucket, ""), "S3 Bucket")
	flag.StringVar(&this.BucketRoot, "s3_bucket_root", EnvString(EnvS3BucketRoot, "_redpill/conf"), "S3 Bucket Root")
	flag.StringVar(&this.AccessKey, "s3_access_key", EnvString(EnvS3AccessKey, ""), "S3 Access Key")
	flag.StringVar(&this.AccessToken, "s3_access_token", EnvString(EnvS3AccessToken, ""), "S3 Access Token")
}
