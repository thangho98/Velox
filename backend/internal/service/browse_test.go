package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowseService_New(t *testing.T) {
	t.Parallel()
	svc := NewBrowseService(nil, nil, nil, nil)
	assert.NotNil(t, svc)
}
