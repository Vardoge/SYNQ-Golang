package synq_golang

import (
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	assert.NotNil(t, new(synq.Api))
	assert.NotNil(t, new(synq.Video))
}
