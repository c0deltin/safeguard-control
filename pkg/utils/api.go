package utils

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

type APIError struct {
	Message string `json:"message"`
}

func Error(logger *logrus.Logger, status int, msg string) events.APIGatewayProxyResponse {
	by, err := json.Marshal(APIError{Message: msg})
	if err != nil {
		logger.Error(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "application/json",
		},
		Body: string(by),
	}
}
