package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"model"
	"net/http"
	"os"
	"utils"

	"codeltin.io/safeguard/control/get-captures/repository"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Lambda struct {
	captureRepository repository.Capture
}

func (l *Lambda) handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	deviceID := r.PathParameters["deviceID"]
	if deviceID == "" {
		log.Println("[ERROR] missing path parameter deviceID")
		return utils.Error(http.StatusBadRequest, "missing parameter \"deviceID\""), nil
	}

	results, err := l.captureRepository.Find(deviceID)
	if err != nil {
		log.Printf("[ERROR] failed to find captures for device %s, err: %v", deviceID, err)
		return utils.Error(http.StatusInternalServerError, err.Error()), nil
	}

	var captures []model.Capture
	for _, x := range results {
		captures = append(captures, *x.MarshalToCapture())
	}

	by, err := json.Marshal(model.CapturesResponse{Captures: captures})
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
		captureRepository: repository.NewCaptureRepository(db, os.Getenv("CONTROL_TABLE_NAME")),
	}

	lambda.Start(l.handler)
}
