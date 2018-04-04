package synq_aws

import (
  "bytes"
  "io/ioutil"
  "net/http"
  "os"
  "testing"

  "github.com/stretchr/testify/require"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/aws/request"
  "github.com/aws/aws-sdk-go/service/dynamodb"
)

func setupDB() *session.Session {
  os.Setenv("AWS_ACCESS_KEY_ID", "derf")
  os.Setenv("AWS_SECRET_ACCESS_KEY", "derf")

  sess := session.New(&aws.Config{Region: aws.String("us-east-2")})
  sess.Handlers.Send.Clear()
  sess.Handlers.Send.PushFront(func(r *request.Request) {
    code    := 403
    resp, _ := ioutil.ReadFile("../sample/aws/error_messages.xml")

    if *r.Params.(*dynamodb.UpdateItemInput).Key["id"].S == "good" {
      code = 200
      resp = []byte("{}")
    }

    r.HTTPResponse = &http.Response{
      StatusCode: code,
      Body:       ioutil.NopCloser(bytes.NewReader(resp)),
      Header:     http.Header{"X-Amzn-Requestid": []string{"123454232"}},
    }
  })

  return sess
}

func TestUpdateRowSuccess(t *testing.T) {
  assert    := require.New(t)
  sess      := setupDB()
  resp, err := LogResult(sess, "good", "200 OK")

  assert.Equal(resp, 200)
  assert.Nil(err)
}

func TestUpdateRowFailure(t *testing.T) {
  assert    := require.New(t)
  sess      := setupDB()
  resp, err := LogResult(sess, "bad", "200 OK")

  assert.Equal(resp, 400)
  assert.NotNil(err)
}