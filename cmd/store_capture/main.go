package main

import (
	"codeltin.io/safeguard/control/store-capture/repository"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elastictranscoder/elastictranscoderiface"
	"log"
	"model"
	"os"
	"time"
	"utils"

	"codeltin.io/safeguard/control/store-capture/bucket"
	"codeltin.io/safeguard/control/store-capture/notifier"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Lambda struct {
	bucket            *bucket.Bucket
	transcoder        elastictranscoderiface.ElasticTranscoderAPI
	captureRepository repository.Capture
	notifier          *notifier.Notifier
}

func (l *Lambda) handler(e events.S3Event) {
	if len(e.Records) == 0 {
		log.Printf("[WARNING] no records for this event")
	}

	record := e.Records[0]

	tags, err := l.bucket.GetObjectTags(record.S3.Object.Key)
	if err != nil {
		log.Printf("[ERROR] failed to get object tags for %s, err: %v", record.S3.Object.Key, err)
		return
	}

	var deviceID *string
	if deviceID = utils.GetTagValue(model.DeviceIDTag, tags); deviceID == nil {
		log.Printf("[ERROR] cannot detect deviceID for object %s", record.S3.Object.Key)
		return
	}

	_, err = l.transcoder.CreateJob(&elastictranscoder.CreateJobInput{
		Input: &elastictranscoder.JobInput{
			Container: aws.String(l.bucket.Name()),
			FrameRate: aws.String("30"),
			Key:       aws.String(record.S3.Object.Key),
		},
		Output: &elastictranscoder.CreateJobOutput{
			PresetId: aws.String("1351620000001-000010"), // Generic 720p
			Key:      aws.String(time.Now().String() + ".mp4"),
		},
		PipelineId: aws.String("safeguard-transcoder"),
	})

	if err != nil {
		log.Printf("[ERROR] failed to create transcoder job, err: %v", err)
		return
	}

	//err = l.captureRepository.Insert(model.CaptureDB{
	//	DeviceID:    *deviceID,
	//	CaptureDate: time.Now().Unix(),
	//	S3ObjectKey: record.S3.Object.Key,
	//})
	//if err != nil {
	//	log.Printf("[ERROR] failed to store capture entry for %s in dynamoDB, err: %v", record.S3.Object.Key, err)
	//	return
	//}
	//
	//r, err := l.notifier.Send()
	//if err != nil {
	//	log.Printf("[ERROR] failed to send notification sms, err: %v", err)
	//	return
	//}

	//log.Printf("[INFO] successfully notified, messageID: %s", *r)
}

func main() {
	s := session.Must(session.NewSession())
	config := aws.NewConfig().
		WithRegion(os.Getenv("AWS_REGION")).
		WithEndpoint(os.Getenv("DYNAMODB_ENDPOINT"))

	db := dynamodbiface.DynamoDBAPI(dynamodb.New(s, config))

	l := &Lambda{
		bucket:            bucket.New(os.Getenv("CAPTURE_BUCKET_NAME"), s, config),
		transcoder:        elastictranscoder.New(s, config.WithRegion("eu-west-1")),
		captureRepository: repository.NewCaptureRepository(db, os.Getenv("CAPTURES_TABLE_NAME")),
		notifier:          notifier.New(s, config).WithPhoneNumber(os.Getenv("SMS_RECEIVER")),
	}

	lambda.Start(l.handler)
}
