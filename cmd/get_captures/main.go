package main

import (
	"encoding/json"
	"model"
	"net/http"
	"os"
	"utils"

	"codeltin.io/safeguard/control/get-captures/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
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

	results, err := l.captureRepository.Find(deviceID)
	if err != nil {
		l.logger.Errorf("failed to find captures for device %s, err: %v", deviceID, err)
		return utils.Error(l.logger, http.StatusInternalServerError, err.Error()), nil
	}

	var captures []model.Capture
	for _, x := range results {
		captures = append(captures, *x.MarshalToRequest())
	}

	// prevent returning "null" in api response
	if captures == nil {
		captures = []model.Capture{}
	}

	by, err := json.Marshal(model.CapturesResponse{Captures: captures})
	if err != nil {
		l.logger.Error(err)
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
