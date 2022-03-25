package repository

import (
	"errors"
	"model"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Capture interface {
	FindOne(deviceID string, date int64) (*model.CaptureDB, error)
}

type captureRepository struct {
	db    dynamodbiface.DynamoDBAPI
	table string
}

func NewCaptureRepository(db dynamodbiface.DynamoDBAPI, table string) *captureRepository {
	return &captureRepository{
		db:    db,
		table: table,
	}
}

var ErrNotFound = errors.New("device_not_found")

func (c *captureRepository) FindOne(deviceID string, date int64) (*model.CaptureDB, error) {
	input := dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deviceID": {
				S: aws.String(deviceID),
			},
			":captureDate": {
				N: aws.String(strconv.FormatInt(date, 10)),
			},
		},
		KeyConditionExpression: aws.String("deviceID = :deviceID AND captureDate = :captureDate"),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		TableName:              aws.String(c.table),
	}

	queryResult, err := c.db.Query(&input)
	if err != nil {
		return nil, err
	}

	if len(queryResult.Items) == 0 {
		return nil, ErrNotFound
	}

	var result model.CaptureDB
	if err = dynamodbattribute.UnmarshalMap(queryResult.Items[0], &result); err != nil {
		return nil, err
	}

	return &result, nil
}
