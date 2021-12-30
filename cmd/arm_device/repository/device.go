package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Device interface {
	UpdateArmedStatus(id string, armed bool) error
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

func (d *deviceRepository) UpdateArmedStatus(id string, armed bool) error {
	input := dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":isArmed": {
				BOOL: aws.Bool(armed),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"deviceID": {
				S: aws.String(id),
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		ReturnValues:           aws.String(dynamodb.ReturnValueNone),
		TableName:              aws.String(d.table),
		UpdateExpression:       aws.String("set isArmed = :isArmed"),
	}

	_, err := d.db.UpdateItem(&input)

	return err
}
