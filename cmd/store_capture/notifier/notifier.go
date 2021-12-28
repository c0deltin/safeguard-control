package notifier

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type Notifier struct {
	svc         *sns.SNS
	phoneNumber string
}

func New(s *session.Session, cfg *aws.Config) *Notifier {
	return &Notifier{
		svc: sns.New(s, cfg),
	}
}

func (n *Notifier) WithPhoneNumber(s string) *Notifier {
	n.phoneNumber = s

	return n
}

func (n *Notifier) PhoneNumber() string {
	return n.phoneNumber
}

func (n *Notifier) Send() (*string, error) {
	input := sns.PublishInput{
		Message:     aws.String("Lorem Ipsum test 123"),
		PhoneNumber: aws.String(n.PhoneNumber()),
	}

	o, err := n.svc.Publish(&input)
	if err != nil {
		return nil, err
	}

	return o.MessageId, nil
}
