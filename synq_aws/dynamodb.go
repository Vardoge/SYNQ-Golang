package synq_aws

import (
  // "errors"
  "time"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
)

// TODO : Figure out how to test this.  AWS does not make this really something they want us to test locally...
func GenerateSession(region string) *session.Session {
  // Just generate a session and return it
  return session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
}

// Writes output to the database, uses current time as the end time
func LogResult(sess *session.Session, id string, response string) (int, error) {
  // create the service object from the session
  svc := dynamodb.New(sess)

  // create the item input object
  // NOTE : Cannot insert items that are not already present!
  input := &dynamodb.UpdateItemInput{
    ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
      ":r":   { S: aws.String(response)             },
      ":n":   { S: aws.String(time.Now().String())  },
      ":id":  { S: aws.String(id)                   }, 
    },
    Key: map[string]*dynamodb.AttributeValue{
      "id": { S: aws.String(id) },
    },
    TableName:            aws.String("message_logs"),
    ReturnValues:         aws.String("ALL_NEW"),
    UpdateExpression:     aws.String("set api_result = :r, end_time = :n"),
    ConditionExpression:  aws.String("id = :id"),
  }

  // determine if there is an error
  _, err := svc.UpdateItem(input)
  if err != nil {
    return 400, err // return error
  }

  // return 200 OK response
  return 200, nil
}