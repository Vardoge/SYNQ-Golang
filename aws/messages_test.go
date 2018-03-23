package aws

import (
  "bytes"
  "io/ioutil"
  "net/http"
  "testing"

  "github.com/stretchr/testify/require"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/aws/request"
  "github.com/aws/aws-sdk-go/service/sqs"
)

// NOTE : aws-sdk-go only understands XML format, JSON format causes a panic error.

func setup() (*session.Session) {
  sess := session.New(&aws.Config{Region: aws.String("us-east-2")})
  sess.Handlers.Send.Clear()
  sess.Handlers.Send.PushFront(func(r *request.Request) {
    var resp []byte
    var code int

    switch *r.Params.(*sqs.ReceiveMessageInput).QueueUrl {
      case "good":
        resp, _ = ioutil.ReadFile("../sample/aws/receive_messages.xml")
        code = 200
      case "empty":
        resp, _ = ioutil.ReadFile("../sample/aws/empty_messages.xml")
        code = 200
      default:
        resp, _ = ioutil.ReadFile("../sample/aws/error_result.xml")
        code = 403
    }


    r.HTTPResponse = &http.Response{
      StatusCode: code,
      Body:       ioutil.NopCloser(bytes.NewReader(resp)),
      Header:     http.Header{"X-Amzn-Requestid": []string{"123454232"}},
    }
  })

  return sess
}

func TestReceiveMessagesSuccess(t *testing.T) {
  assert    := require.New(t)
  sess      := setup()
  resp, err := ReceiveMessages(sess, "good")

  assert.Nil(err)
  assert.Equal(len(resp), 1)
  assert.Equal(resp[0], "This is a test message")
}

func TestReceiveMessagesEmptyList(t *testing.T) {
  assert    := require.New(t)
  sess      := setup()
  resp, err := ReceiveMessages(sess, "empty")

  assert.Nil(err)
  assert.Equal(len(resp), 0)
}

func TestReceiveMessagesError(t *testing.T) {
  assert    := require.New(t)
  sess      := setup()
  resp, err := ReceiveMessages(sess, "forbidden")

  assert.NotNil(err)
  assert.Equal(len(resp), 0)
}