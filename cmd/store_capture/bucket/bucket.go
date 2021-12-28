package bucket

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Bucket struct {
	name string
	svc  *s3.S3
}

func New(name string, s *session.Session, cfg *aws.Config) *Bucket {
	return &Bucket{
		name: name,
		svc:  s3.New(s, cfg),
	}
}

func (b *Bucket) Name() string {
	return b.name
}

func (b *Bucket) GetObjectTags(key string) ([]*s3.Tag, error) {
	input := s3.GetObjectTaggingInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}

	o, err := b.svc.GetObjectTagging(&input)
	if err != nil {
		return nil, err
	}

	return o.TagSet, nil
}
