package repository

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Capture interface {
	Find(deviceID string) ([]model.CaptureDB, error)
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

func (c *captureRepository) Find(deviceID string) ([]model.CaptureDB, error) {
	input := dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deviceID": {
				S: aws.String(deviceID),
			},
		},
		KeyConditionExpression: aws.String("deviceID = :deviceID"),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		TableName:              aws.String(c.table),
	}

	queryResult, err := c.db.Query(&input)
	if err != nil {
		return nil, err
	}

	var result []model.CaptureDB
	if err = dynamodbattribute.UnmarshalListOfMaps(queryResult.Items, &result); err != nil {
		return nil, err
	}

	return result, nil
}
