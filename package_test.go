package synq_golang

import (
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	assert.NotNil(t, new(synq.ApiV2))
	assert.NotNil(t, new(synq.VideoV2))
}
