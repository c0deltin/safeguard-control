package main

import (
	"encoding/json"
	"model"
	"net/http"
	"os"
	"utils"

	"github.com/c0deltin/safeguard-control/get-devices/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	deviceRepository repository.Device
	logger           *logrus.Logger
}

func (l *Lambda) handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	devices, err := l.deviceRepository.Find()
	if err != nil {
		var status = http.StatusInternalServerError

		l.logger.Errorf("failed to fetch devices, err: %v", err)
		return utils.Error(l.logger, status, err.Error()), nil
	}

	by, err := json.Marshal(model.DevicesResponse{Devices: model.ConvertSliceToRequest(devices)})
	if err != nil {
		l.logger.Errorf("failed to marshal response objects, err: %v", err)
		return utils.Error(l.logger, http.StatusInternalServerError, err.Error()), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(by),
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
	}

	l.logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(l.handler)
}
