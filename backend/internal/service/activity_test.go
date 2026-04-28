package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thawng/velox/internal/repository"
)

func TestActivityConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 256, activityBufferSize)
	assert.Equal(t, 50, activityFlushSize)
	assert.Equal(t, 5*time.Second, activityFlushEvery)
}

func TestActivityService_Log_DoesNotBlock(t *testing.T) {
	t.Parallel()

	svc := &ActivityService{
		ch:   make(chan repository.ActivityEntry, 10),
		done: make(chan struct{}),
		wg:   sync.WaitGroup{},
	}
	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		for {
			select {
			case <-svc.ch:
			case <-svc.done:
				return
			}
		}
	}()

	now := time.Now()
	var userID int64 = 1
	svc.Log(&userID, "test_action", "127.0.0.1", nil, "{}")
	elapsed := time.Since(now)

	assert.Less(t, int(elapsed.Milliseconds()), 100)
	close(svc.done)
	svc.wg.Wait()
}

func TestActivityService_Log_AllFields(t *testing.T) {
	t.Parallel()

	svc := &ActivityService{
		ch:   make(chan repository.ActivityEntry, 10),
		done: make(chan struct{}),
		wg:   sync.WaitGroup{},
	}
	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		for {
			select {
			case entry := <-svc.ch:
				assert.Equal(t, int64(123), *entry.UserID)
				assert.Equal(t, "playback", entry.Action)
				assert.Equal(t, "192.168.1.1", entry.IP)
				assert.Equal(t, int64(456), *entry.MediaID)
				assert.Equal(t, `{"position":120}`, entry.Details)
			case <-svc.done:
				return
			}
		}
	}()

	var userID int64 = 123
	var mediaID int64 = 456
	svc.Log(&userID, "playback", "192.168.1.1", &mediaID, `{"position":120}`)

	time.Sleep(50 * time.Millisecond)
	close(svc.done)
	svc.wg.Wait()
}

func TestActivityService_Close_Waits(t *testing.T) {
	t.Parallel()

	svc := &ActivityService{
		ch:   make(chan repository.ActivityEntry, 256),
		done: make(chan struct{}),
		wg:   sync.WaitGroup{},
	}
	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		<-svc.done
	}()

	close(svc.done)
	svc.wg.Wait()
}
