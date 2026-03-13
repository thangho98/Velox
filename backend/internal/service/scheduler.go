package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
)

// Task represents a scheduled job.
type Task struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
	LastRun  time.Time
	NextRun  time.Time
	Running  bool
}

// Scheduler manages periodic tasks.
type Scheduler struct {
	mu    sync.Mutex
	tasks map[string]*Task
	done  chan struct{}
	wg    sync.WaitGroup
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
		done:  make(chan struct{}),
	}
}

// Register adds a task to the scheduler. Must be called before Start.
func (s *Scheduler) Register(name string, interval time.Duration, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = &Task{
		Name:     name,
		Interval: interval,
		Fn:       fn,
		NextRun:  time.Now().Add(interval),
	}
}

// Start begins the scheduler loop, checking every minute for tasks that need to run.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.loop()
	log.Printf("scheduler: started with %d tasks", len(s.tasks))
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	close(s.done)
	s.wg.Wait()
	log.Println("scheduler: stopped")
}

// ListTasks returns info about all registered tasks.
func (s *Scheduler) ListTasks() []model.TaskInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var infos []model.TaskInfo
	for _, t := range s.tasks {
		info := model.TaskInfo{
			Name:     t.Name,
			Interval: t.Interval.String(),
			NextRun:  t.NextRun.Format(time.RFC3339),
			Running:  t.Running,
		}
		if !t.LastRun.IsZero() {
			info.LastRun = t.LastRun.Format(time.RFC3339)
		}
		infos = append(infos, info)
	}
	return infos
}

// RunNow triggers a task immediately by name.
func (s *Scheduler) RunNow(name string) error {
	s.mu.Lock()
	task, ok := s.tasks[name]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("task %q not found", name)
	}
	if task.Running {
		s.mu.Unlock()
		return fmt.Errorf("task %q is already running", name)
	}
	s.mu.Unlock()

	go s.runTask(task)
	return nil
}

func (s *Scheduler) loop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.done:
			return
		}
	}
}

func (s *Scheduler) tick() {
	now := time.Now()

	s.mu.Lock()
	var toRun []*Task
	for _, t := range s.tasks {
		if !t.Running && now.After(t.NextRun) {
			toRun = append(toRun, t)
		}
	}
	s.mu.Unlock()

	for _, t := range toRun {
		go s.runTask(t)
	}
}

func (s *Scheduler) runTask(t *Task) {
	s.mu.Lock()
	if t.Running {
		s.mu.Unlock()
		return
	}
	t.Running = true
	s.mu.Unlock()

	log.Printf("scheduler: running task %q", t.Name)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	err := t.Fn(ctx)

	elapsed := time.Since(start)

	s.mu.Lock()
	t.Running = false
	t.LastRun = time.Now()
	t.NextRun = time.Now().Add(t.Interval)
	s.mu.Unlock()

	if err != nil {
		log.Printf("scheduler: task %q failed after %s: %v", t.Name, elapsed, err)
	} else {
		log.Printf("scheduler: task %q completed in %s", t.Name, elapsed)
	}
}
