package aws

import (
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/sqs"
)

func ReceiveMessages(url string, region string) ([]string, error) {
  // Create the session, assumes credentials are provided elsewhere
  sess  :=  session.Must(session.NewSession(&aws.Config{  Region: aws.String(region)  }))
  svc   :=  sqs.New(sess)

  // Setup SQS Parameters
  // NOTE : VisibilityTimeout keeps us from getting the same message repeatedly
  receiveParams := &sqs.ReceiveMessageInput{
    QueueUrl:             aws.String(url),
    MaxNumberOfMessages:  aws.Int64(10),
    VisibilityTimeout:    aws.Int64(30),
  }

  // get the messages
  resp, err := svc.ReceiveMessage(receiveParams)
  if err != nil {
    return nil, err
  }

  // pull the message body
  var messages []string
  for _, element := range resp.Messages {
    messages = append(messages, *element.Body)
  }

  // return the messages
  return messages, nil
}