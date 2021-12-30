package main

import (
	"encoding/json"
	"errors"
	"log"
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
)

type Lambda struct {
	deviceRepository repository.Device
}

func (l *Lambda) handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	deviceID := r.PathParameters["deviceID"]
	if deviceID == "" {
		log.Println("[ERROR] missing path parameter deviceID")
		return utils.Error(http.StatusBadRequest, "missing parameter \"deviceID\""), nil
	}

	device, err := l.deviceRepository.FindOne(deviceID)
	if err != nil {
		var status = http.StatusInternalServerError
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}

		log.Printf("[ERROR] failed to find device %s, err: %v", deviceID, err)
		return utils.Error(status, err.Error()), nil
	}

	by, err := json.Marshal(model.DeviceResponse{Device: device.MarshalToRequest()})
	if err != nil {
		log.Printf("[ERROR] failed to marshal response object, err: %v", err)
		return utils.Error(http.StatusInternalServerError, err.Error()), nil
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
	}

	lambda.Start(l.handler)
}
