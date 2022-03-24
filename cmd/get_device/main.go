package main

import (
	"encoding/json"
	"errors"
	"model"
	"net/http"
	"os"
	"utils"

	"codeltin.io/safeguard/control/get-device/repository"

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
	deviceID := r.PathParameters["deviceID"]
	if deviceID == "" {
		l.logger.Errorf("missing path parameter %s", "deviceID")
		return utils.Error(l.logger, http.StatusBadRequest, "missing parameter \"deviceID\""), nil
	}

	device, err := l.deviceRepository.FindOne(deviceID)
	if err != nil {
		var status = http.StatusInternalServerError
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}

		l.logger.Errorf("failed to find device %s, err: %v", deviceID, err)
		return utils.Error(l.logger, status, err.Error()), nil
	}

	by, err := json.Marshal(model.DeviceResponse{Device: device.MarshalToRequest()})
	if err != nil {
		l.logger.Errorf("failed to marshal response object, err: %v", err)
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
