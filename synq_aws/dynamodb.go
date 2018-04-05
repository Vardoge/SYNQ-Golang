package synq_aws

import (
  // "errors"
  "time"

  "github.com/google/uuid"
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

// create a log line
func CreateLog(sess *session.Session, worker string, token string, datatype string, data string) (string, error) {
  // generate the UUID and get the service
  id  := uuid.New().String()
  svc := dynamodb.New(sess)

  // build the inupts
  input := &dynamodb.PutItemInput{
    Item: map[string]*dynamodb.AttributeValue{
      "id":           { S: aws.String(id)                   },
      "start_time":   { S: aws.String(time.Now().String())  },
      "end_time":     { S: aws.String("in progress")        },
      "token":        { S: aws.String(token)                },
      "datatype":     { S: aws.String(datatype)             },
      "data":         { S: aws.String(data)                 },
      "worker":       { S: aws.String(worker)               },
      "api_response": { S: aws.String("in progress")        },
    },
    ReturnConsumedCapacity: aws.String("TOTAL"),
    TableName:              aws.String("message_logs"),
  }

  // Insert the item into the db
  _, err := svc.PutItem(input)
  if err != nil {
    // if there was an error return the error
    return "ERROR", err
  }

  // return the id and success if no error
  return id, nil
} 