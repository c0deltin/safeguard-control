package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"model"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Capture interface {
	Insert(capture model.CaptureDB) error
	CountByObjectKey(deviceID, key string) (*int64, error)
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

func (c *captureRepository) CountByObjectKey(deviceID, key string) (*int64, error) {
	input := dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deviceID": {
				S: aws.String(deviceID),
			},
			":captureDate": {
				N: aws.String(strconv.FormatInt(time.Now().UnixMilli(), 10)),
			},
			":s3ObjectKey": {
				S: aws.String(key),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#s3ObjectKey": aws.String("s3ObjectKey"),
		},
		KeyConditionExpression: aws.String("deviceID = :deviceID AND captureDate < :captureDate"),
		FilterExpression:       aws.String("#s3ObjectKey = :s3ObjectKey"),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityNone),
		TableName:              aws.String(c.table),
	}

	result, err := c.db.Query(&input)
	if err != nil {
		return nil, err
	}

	return result.Count, nil
}
