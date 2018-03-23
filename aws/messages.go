package synq_aws

import (
  "strings"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/sqs"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type SQSMessage struct {
  Id      string  `json:"message_id"`
  Handle  string  `json:"-"`
  Body    string  `json:"message"`
  URL     string  `json:"service_url"`
  Result  string  `json:"result"`
}

// TODO : Figure out how to test this.  AWS does not make this really something they want us to test locally...
func GenerateSession(region string) *session.Session {
  // Just generate a session and return it
  return session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-2")}))
}

func ReceiveMessages(sess *session.Session, url string) ([]SQSMessage, error) {
  // Create the session, assumes credentials are provided elsewhere
  svc := sqs.New(sess)

  // Setup SQS Parameters
  // NOTE : VisibilityTimeout keeps us from getting the same message repeatedly
  receiveParams := &sqs.ReceiveMessageInput{
    QueueUrl:             aws.String(url),
    MaxNumberOfMessages:  aws.Int64(10),
    VisibilityTimeout:    aws.Int64(10),
  }

  // get the messages
  resp, err := svc.ReceiveMessage(receiveParams)
  if err != nil {
    return nil, err
  }

  // pull the message body
  var messages []SQSMessage
  for _, element := range resp.Messages {
    messages = append(messages, SQSMessage{
      Id:     strings.TrimSpace(*element.MessageId),
      Body:   strings.TrimSpace(*element.Body),
      Handle: strings.TrimSpace(*element.ReceiptHandle),
      URL:    url })
  }

  // return the messages
  return messages, nil
}

func ResolveMessage(sess *session.Session, msg SQSMessage) (int, error) {
  // setup the DB service and convert the message to a DynamoDB item
  dbSVC := dynamodb.New(sess)
  av, createErr := dynamodbattribute.MarshalMap(msg)
  if createErr != nil {
    return 400, createErr
  }

  // create the item for the database
  input := &dynamodb.PutItemInput{
    Item:       av,
    TableName:  aws.String("sqs_message_results"),
  }

  // insert the item into the database
  _, putErr := dbSVC.PutItem(input)
  if putErr != nil {
    return 400, putErr
  }

  // Create the SQS service
  sqsSVC := sqs.New(sess)

  // Setup the parameters
  deleteParams := &sqs.DeleteMessageInput{
    QueueUrl:       aws.String(msg.URL),
    ReceiptHandle:  aws.String(msg.Handle),
  }

  // Send the message and return an error if there is one
  _, delErr := sqsSVC.DeleteMessage(deleteParams)
  if delErr != nil {
    return 400, delErr
  }

  return 204, nil
}