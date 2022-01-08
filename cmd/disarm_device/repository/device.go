package repository

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Device interface {
	UpdateArmedStatus(id string, armed bool) (*model.DeviceDB, error)
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

func (d *deviceRepository) UpdateArmedStatus(id string, armed bool) (*model.DeviceDB, error) {
	input := dynamodb.UpdateItemInput{
		ConditionExpression: aws.String("attribute_exists(id)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":isArmed": {
				BOOL: aws.Bool(armed),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		ReturnValues:           aws.String(dynamodb.ReturnValueAllNew),
		TableName:              aws.String(d.table),
		UpdateExpression:       aws.String("set isArmed = :isArmed"),
	}

	result, err := d.db.UpdateItem(&input)
	if err != nil {
		return nil, err
	}

	var device model.DeviceDB
	if err = dynamodbattribute.UnmarshalMap(result.Attributes, &device); err != nil {
		return nil, err
	}

	return &device, nil
}
