package main

import (
	"model"
	"os"
	"time"
	"utils"

	"codeltin.io/safeguard/control/store-capture/bucket"
	"codeltin.io/safeguard/control/store-capture/notifier"
	"codeltin.io/safeguard/control/store-capture/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	bucket            *bucket.Bucket
	captureRepository repository.Capture
	logger            *logrus.Logger
	notifier          *notifier.Notifier
}

func (l *Lambda) handler(e events.S3Event) {
	if len(e.Records) == 0 {
		l.logger.Warnln("no records for this event")
	}

	recordKey := e.Records[0].S3.Object.Key

	tags, err := l.bucket.GetObjectTags(recordKey)
	if err != nil {
		l.logger.Errorf("failed to get object tags for %s, err: %v", recordKey, err)
		return
	}

	var deviceID *string
	if deviceID = utils.GetTagValue(model.DeviceIDTag, tags); deviceID == nil {
		l.logger.Errorf("cannot detect deviceID for object %s", recordKey)
		return
	}

	count, err := l.captureRepository.CountByObjectKey(*deviceID, recordKey)
	if err != nil || *count > 0 {
		l.logger.Error("query failed or item already exists (deviceID: %s, key: %s, err: %v)", *deviceID, recordKey, err)
		return
	}

	err = l.captureRepository.Insert(model.CaptureDB{
		DeviceID:    *deviceID,
		CaptureDate: time.Now().Unix(),
		S3ObjectKey: recordKey,
	})
	if err != nil {
		l.logger.Errorf("failed to store capture entry for %s in dynamoDB, err: %v", recordKey, err)
		return
	}

	r, err := l.notifier.Send()
	if err != nil {
		l.logger.Errorf("failed to send notification sms, err: %v", err)
		return
	}

	l.logger.Infof("successfully notified, messageID: %s", *r)
}

func main() {
	s := session.Must(session.NewSession())
	config := aws.NewConfig().
		WithRegion(os.Getenv("AWS_REGION")).
		WithEndpoint(os.Getenv("DYNAMODB_ENDPOINT"))

	db := dynamodbiface.DynamoDBAPI(dynamodb.New(s, config))

	l := &Lambda{
		bucket:            bucket.New(os.Getenv("CAPTURE_BUCKET_NAME"), s, config),
		captureRepository: repository.NewCaptureRepository(db, os.Getenv("CAPTURES_TABLE_NAME")),
		logger:            logrus.New(),
		notifier:          notifier.New(s, config).WithPhoneNumber(os.Getenv("SMS_RECEIVER")),
	}

	l.logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(l.handler)
}
