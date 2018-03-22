package aws

import (
  "testing"

  "github.com/stretchr/testify/require"
)

func TestReceiveMessages(t *testing.T) {
  assert := require.New(t)
  url := "https://sqs.us-east-2.amazonaws.com/072327369740/metadata"

  resp, err := ReceiveMessages(url, "us-east-2")

  assert.Nil(err)
  assert.Equal(len(resp), 0)
}