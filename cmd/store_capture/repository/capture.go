package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"model"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Capture interface {
	Insert(capture model.CaptureDB) error
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

func (c *captureRepository) Insert(capture model.CaptureDB) error {
	m, err := dynamodbattribute.MarshalMap(capture)
	if err != nil {
		return err
	}

	input := dynamodb.PutItemInput{
		Item:                   m,
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		ReturnValues:           aws.String(dynamodb.ReturnValueNone),
		TableName:              aws.String(c.table),
	}

	_, err = c.db.PutItem(&input)

	return err
}
