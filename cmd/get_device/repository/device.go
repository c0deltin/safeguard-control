package repository

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"model"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Device interface {
	FindOne(id string) (*model.DeviceDB, error)
}

type deviceRepository struct {
	db    dynamodbiface.DynamoDBAPI
	table string
}

func NewDeviceRepository(db dynamodbiface.DynamoDBAPI, table string) *deviceRepository {
	return &deviceRepository{
		db:    db,
		table: table,
	}
}

var ErrNotFound = errors.New("device_not_found")

func (d *deviceRepository) FindOne(id string) (*model.DeviceDB, error) {
	input := dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deviceID": {
				S: aws.String(id),
			},
		},
		KeyConditionExpression: aws.String("deviceID = :deviceID"),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		TableName:              aws.String(d.table),
	}

	result, err := d.db.Query(&input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, ErrNotFound
	}

	var device model.DeviceDB
	if err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &device); err != nil {
		return nil, err
	}

	return &device, nil
}
