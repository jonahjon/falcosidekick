package awsutil

import (
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// "us-east-1 is used because it's where AWS first provides support for new features,
const DefaultRegion = "us-east-1"

// This is nil by default, but is exposed in case it needs to be changed for tests.
var ec2Endpoint *string

/*
Our chosen approach is:
	"More specific takes precedence over less specific."
1. User-provided configuration is the most explicit.
2. Environment variables are potentially shared across many invocations and so they have less precedence.
3. Configuration in `~/.aws/config` is shared across all invocations of a given user and so this has even less precedence.
4. Configuration retrieved from the EC2 instance metadata service is shared by all invocations on a given machine, and so it has the lowest precedence.
This approach should be used in future updates to this logic.
*/
func GetRegion(configuredRegion string) (string, error) {
	if configuredRegion != "" {
		return configuredRegion, nil
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return "", errors.Wrap(err, "got error when starting session")
	}

	region := aws.StringValue(sess.Config.Region)
	if region != "" {
		return region, nil
	}

	metadata := ec2metadata.New(sess, &aws.Config{
		Endpoint:                          ec2Endpoint,
		EC2MetadataDisableTimeoutOverride: aws.Bool(true),
		HTTPClient: &http.Client{
			Timeout: time.Second,
		},
	})
	if !metadata.Available() {
		return DefaultRegion, nil
	}

	region, err = metadata.Region()
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve region from instance metadata")
	}
	return region, nil
}
