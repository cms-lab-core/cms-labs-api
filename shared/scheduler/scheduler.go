package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	"github.com/rs/zerolog"
)

// Scheduler структура для управления задачами
type Scheduler struct {
	tasks    []Task
	wg       sync.WaitGroup
	done     chan bool
	Interval time.Duration
	*zerolog.Logger
}

// NewScheduler создает новый экземпляр Scheduler
func NewScheduler(Interval time.Duration) *Scheduler {
	return &Scheduler{
		done:     make(chan bool),
		Interval: Interval,
		Logger:   logs.NewZeroLogger(logs.NewZeroLoggerConf(nil).SetName("scheduler")),
	}
}

// AddTask добавляет задачу в планировщик
func (s *Scheduler) AddTask(task Task) *Scheduler {
	s.tasks = append(s.tasks, task)
	return s
}

// Start запускает выполнение задач с указанным интервалом
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for _, task := range s.tasks {
			if err := task.Execute(); err != nil {
				s.Logger.Warn().Msg(fmt.Sprintf("Task execution error: %v\n", err))
			}
		}
		ticker := time.NewTicker(s.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for _, task := range s.tasks {
					if err := task.Execute(); err != nil {
						s.Logger.Warn().Msg(fmt.Sprintf("Task execution error: %v\n", err))
					}
				}
			case <-s.done:
				return
			}
		}
	}()
}

// Stop останавливает выполнение задач
func (s *Scheduler) Stop() {
	close(s.done)
	s.wg.Wait()
}
