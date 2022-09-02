package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"model"
	"net/http"
	"os"
	"strconv"
	"utils"

	"github.com/c0deltin/safeguard-control/get-capture/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	captureRepository repository.Capture
	logger            *logrus.Logger
}

func (l *Lambda) handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	deviceID := r.PathParameters["deviceID"]
	if deviceID == "" {
		l.logger.Errorf("missing path parameter %s", "deviceID")
		return utils.Error(l.logger, http.StatusBadRequest, "missing parameter \"deviceID\""), nil
	}

	captureDateStr := r.PathParameters["captureDate"]
	if deviceID == "" {
		l.logger.Errorf("missing path parameter %s", "captureDate")
		return utils.Error(l.logger, http.StatusBadRequest, "missing parameter \"captureDate\""), nil
	}

	captureDate, err := strconv.ParseInt(captureDateStr, 10, 64)
	if err != nil {
		l.logger.Errorf("invalid parameter %s", "captureDate")
		return utils.Error(l.logger, http.StatusBadRequest, "invalid parameter \"captureDate\""), nil
	}

	capture, err := l.captureRepository.FindOne(deviceID, captureDate)
	if err != nil {
		var status = http.StatusInternalServerError
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}

		l.logger.Errorf("failed to find capture %s on %d, err: %v", deviceID, captureDate, err)
		return utils.Error(l.logger, status, err.Error()), nil
	}

	by, err := json.Marshal(model.CaptureResponse{Capture: capture.ConvertToRequest()})
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
		captureRepository: repository.NewCaptureRepository(db, os.Getenv("CAPTURES_TABLE_NAME")),
		logger:            logrus.New(),
	}

	l.logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(l.handler)
}
