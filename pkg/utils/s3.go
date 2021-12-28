package utils

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetTagValue(key string, tags []*s3.Tag) *string {
	for _, tag := range tags {
		if *tag.Key == key {
			return tag.Value
		}
	}

	return nil
}
