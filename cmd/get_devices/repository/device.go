package repository

import (
	"errors"
	"model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Device interface {
	Find() ([]model.DeviceDB, error)
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

func (d *deviceRepository) Find() ([]model.DeviceDB, error) {
	input := dynamodb.ScanInput{
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		TableName:              aws.String(d.table),
	}

	result, err := d.db.Scan(&input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var device []model.DeviceDB
	if err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &device); err != nil {
		return nil, err
	}

	return device, nil
}
