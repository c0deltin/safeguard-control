package main

import (
	"encoding/json"
	"net/http"
	"os"
	"utils"

	"codeltin.io/safeguard/control/arm-device/notifier"
	"codeltin.io/safeguard/control/arm-device/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	deviceRepository repository.Device
	logger           *logrus.Logger
	notifier         *notifier.Notifier
}

func (l *Lambda) handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	deviceID := r.PathParameters["deviceID"]
	if deviceID == "" {
		l.logger.Errorf("missing path parameter %s", "deviceID")
		return utils.Error(l.logger, http.StatusBadRequest, "missing parameter \"deviceID\""), nil
	}

	result, err := l.deviceRepository.UpdateArmedStatus(deviceID, true)
	if err != nil {
		var status = http.StatusInternalServerError
		if aErr, ok := err.(awserr.Error); ok {
			switch aErr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException, dynamodb.ErrCodeConditionalCheckFailedException:
				status = http.StatusNotFound
			}
		}
		l.logger.Errorf("failed to arm device %s: %v", deviceID, err)
		return utils.Error(l.logger, status, err.Error()), nil
	}

	by, err := json.Marshal(result.MarshalToRequest())
	if err != nil {
		l.logger.Error(err)
		return utils.Error(l.logger, http.StatusInternalServerError, err.Error()), nil
	}

	msg, err := l.notifier.Send(deviceID, string(by))
	if err != nil {
		l.logger.Errorf("failed to send sns: %v", err)
		return utils.Error(l.logger, http.StatusInternalServerError, err.Error()), nil
	}

	l.logger.Infof("successfully placed message %s on topic", *msg)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "{}",
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	s := session.Must(session.NewSession())
	config := aws.NewConfig().
		WithRegion(os.Getenv("AWS_REGION")).
		WithEndpoint(os.Getenv("DYNAMODB_ENDPOINT"))

	db := dynamodbiface.DynamoDBAPI(dynamodb.New(s, config))

	l := &Lambda{
		deviceRepository: repository.NewDeviceRepository(db, os.Getenv("DEVICE_TABLE_NAME")),
		logger:           logrus.New(),
		notifier:         notifier.New(s, config).WithTopicArn(os.Getenv("SNS_TOPIC_ARN")),
	}

	l.logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(l.handler)
}
