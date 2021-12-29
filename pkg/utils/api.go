package utils

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
)

type APIError struct {
	Message string `json:"message"`
}

func Error(status int, msg string) events.APIGatewayProxyResponse {
	by, err := json.Marshal(APIError{Message: msg})
	if err != nil {
		log.Printf("[ERROR] failed to marshal error, err: %v", err)
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
