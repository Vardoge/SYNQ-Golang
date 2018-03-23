package aws

import (
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/sqs"
)

// TODO : Figure out how to test this.  AWS does not make this really something they want us to test locally...
func GenerateSession(region string) *session.Session {
  // Just generate a session and return it
  return session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-2")}))
}

func ReceiveMessages(sess *session.Session, url string) ([]string, error) {
  // Create the session, assumes credentials are provided elsewhere
  svc   :=  sqs.New(sess)

  // Setup SQS Parameters
  // NOTE : VisibilityTimeout keeps us from getting the same message repeatedly
  receiveParams := &sqs.ReceiveMessageInput{
    QueueUrl:             aws.String(url),
    MaxNumberOfMessages:  aws.Int64(10),
    VisibilityTimeout:    aws.Int64(600),
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