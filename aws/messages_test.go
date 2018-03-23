package synq_aws

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

func setupReceive() (*session.Session) {
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
  sess      := setupReceive()
  resp, err := ReceiveMessages(sess, "good")

  assert.Nil(err)
  assert.Equal(len(resp), 1)
  assert.Equal(resp[0].Handle, "MbZj6wDWli+JvwwJaBV+3dcjk2YW2vA3+STFFljTM8tJJg6HRG6PYSasuWXPJB+CwLj1FjgXUv1uSj1gUPAWV66FU/WeR4mq2OKpEGYWbnLmpRCJVAyeMjeU5ZBdtcQ+QEauMZc8ZRv37sIW2iJKq3M9MFx1YvV11A2x/KSbkJ0=")
  assert.Equal(resp[0].Body, "This is a test message")
}

func TestReceiveMessagesEmptyList(t *testing.T) {
  assert    := require.New(t)
  sess      := setupReceive()
  resp, err := ReceiveMessages(sess, "empty")

  assert.Nil(err)
  assert.Equal(len(resp), 0)
}

func TestReceiveMessagesError(t *testing.T) {
  assert    := require.New(t)
  sess      := setupReceive()
  resp, err := ReceiveMessages(sess, "forbidden")

  assert.NotNil(err)
  assert.Equal(len(resp), 0)
}

// TODO : replace with actual testing for the DB/SQS process
func TestProcess(t *testing.T) {
  sess := GenerateSession("us-east-2")
  msgs, _ := ReceiveMessages(sess, "https://sqs.us-east-2.amazonaws.com/072327369740/metadata")

  if len(msgs) > 0 {
    msg := msgs[0]
    msg.Result = "200 OK"

    ResolveMessage(sess, msg)
  }
}