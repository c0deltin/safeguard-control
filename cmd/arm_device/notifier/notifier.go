package notifier

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type Notifier struct {
	svc      *sns.SNS
	topicArn string
}

func New(s *session.Session, cfg *aws.Config) *Notifier {
	return &Notifier{
		svc: sns.New(s, cfg),
	}
}

func (n *Notifier) WithTopicArn(t string) *Notifier {
	n.topicArn = t

	return n
}

func (n *Notifier) Send(msg string) (*string, error) {
	input := sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(n.topicArn),
	}

	o, err := n.svc.Publish(&input)
	if err != nil {
		return nil, err
	}

	return o.MessageId, nil
}
