package synq_aws

import (
  "strings"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/sqs"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// TODO : Figure out how to test this.  AWS does not make this really something they want us to test locally...
func GenerateSession(region string) *session.Session {
  // Just generate a session and return it
  return session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-2")}))
}

func CreateEntry(sess *session.Session, msg SQSMessage) (int, error) {
  // setup the DB service and convert the message to a DynamoDB item
  dbSVC := dynamodb.New(sess)
  av, _ := dynamodbattribute.MarshalMap(msg)

  // create the item for the database
  input := &dynamodb.PutItemInput{
    Item:       av,
    TableName:  aws.String("message_results"),
  }

  // insert the item into the database
  _, putErr := dbSVC.PutItem(input)
  if putErr != nil {
    return 400, putErr
  }

  return 200, nil
}