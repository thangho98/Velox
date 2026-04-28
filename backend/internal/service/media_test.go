package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaService_New(t *testing.T) {
	t.Parallel()
	svc := NewMediaService(nil, nil)
	assert.NotNil(t, svc)
}

func TestMediaService_SetterMethods(t *testing.T) {
	t.Parallel()
	svc := NewMediaService(nil, nil)

	t.Run("SetEpisodeRepo_does_not_panic", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			svc.SetEpisodeRepo(nil)
		})
	})

	t.Run("SetSeasonRepo_does_not_panic", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			svc.SetSeasonRepo(nil)
		})
	})

	t.Run("SetImageMetadataRepo_does_not_panic", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			svc.SetImageMetadataRepo(nil)
		})
	})
}
