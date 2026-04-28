package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadService_New(t *testing.T) {
	t.Parallel()
	svc := NewDownloadService(nil, nil, nil, "/tmp")
	assert.NotNil(t, svc)
}

func TestDownloadService_ListTasks(t *testing.T) {
	t.Parallel()
	svc := NewDownloadService(nil, nil, nil, "/tmp")

	tasks := svc.ListTasks()
	assert.NotNil(t, tasks)
	assert.Len(t, tasks, 0)
}
