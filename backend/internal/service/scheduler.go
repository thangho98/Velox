package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
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
	repo  *repository.ScheduledTaskRepository
	done  chan struct{}
	wg    sync.WaitGroup
}

func NewScheduler(repo *repository.ScheduledTaskRepository) *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
		repo:  repo,
		done:  make(chan struct{}),
	}
}

// Register adds a task to the scheduler. Must be called before Start.
func (s *Scheduler) Register(name string, defaultInterval time.Duration, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Default initialization
	nextRun := time.Now().Add(defaultInterval)
	var lastRun time.Time
	interval := defaultInterval

	// Try loading from DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if dbTask, err := s.repo.GetByName(ctx, name); err == nil && dbTask != nil {
		if dbTask.IntervalSeconds > 0 {
			interval = time.Duration(dbTask.IntervalSeconds) * time.Second
		}
		if dbTask.LastRun != nil && !dbTask.LastRun.IsZero() {
			lastRun = *dbTask.LastRun
			// Calculate next run precisely from last run, unless it's already in the past
			potentialNext := lastRun.Add(interval)
			if potentialNext.After(time.Now()) {
				nextRun = potentialNext
			} else {
				// We missed runs, run it soon
				nextRun = time.Now().Add(time.Minute)
			}
		}
	} else if err != nil {
		log.Printf("scheduler: failed to load %q from db: %v", name, err)
	}

	s.tasks[name] = &Task{
		Name:     name,
		Interval: interval,
		Fn:       fn,
		LastRun:  lastRun,
		NextRun:  nextRun,
	}

	// Upsert initial/current state back to DB ensuring it gets tracked
	go func() {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		var lst *time.Time
		if !lastRun.IsZero() {
			lst = &lastRun
		}
		_ = s.repo.Upsert(ctx2, &repository.ScheduledTask{
			Name:            name,
			LastRun:         lst,
			IntervalSeconds: int(interval.Seconds()),
		})
	}()
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

// UpdateInterval changes a task's interval and persists it.
func (s *Scheduler) UpdateInterval(ctx context.Context, name string, interval time.Duration) error {
	s.mu.Lock()
	task, ok := s.tasks[name]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("task %q not found", name)
	}

	task.Interval = interval
	if !task.Running {
		// Recalculate NextRun based on the new interval and LastRun
		if !task.LastRun.IsZero() {
			task.NextRun = task.LastRun.Add(interval)
			if task.NextRun.Before(time.Now()) {
				task.NextRun = time.Now().Add(10 * time.Second)
			}
		} else {
			task.NextRun = time.Now().Add(interval)
		}
	}
	s.mu.Unlock()

	return s.repo.UpdateInterval(ctx, name, int(interval.Seconds()))
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
	now := time.Now()

	s.mu.Lock()
	t.Running = false
	t.LastRun = now
	t.NextRun = now.Add(t.Interval)
	s.mu.Unlock()

	// Persist to DB
	go func() {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		_ = s.repo.UpdateLastRun(ctx2, t.Name, now)
	}()

	if err != nil {
		log.Printf("scheduler: task %q failed after %s: %v", t.Name, elapsed, err)
	} else {
		log.Printf("scheduler: task %q completed in %s", t.Name, elapsed)
	}
}
