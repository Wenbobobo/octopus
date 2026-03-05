package common

import (
	"runtime/debug"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type WorkerPool struct {
	tasks  chan func()
	once   sync.Once
	wg     sync.WaitGroup
	closed atomic.Bool
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = workers
	}

	p := &WorkerPool{
		tasks: make(chan func(), queueSize),
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for task := range p.tasks {
				if task != nil {
					func() {
						defer func() {
							if panicErr := recover(); panicErr != nil {
								log.Errorf("worker task panic: %v\n%s", panicErr, debug.Stack())
							}
						}()
						task()
					}()
				}
			}
		}()
	}

	return p
}

func (p *WorkerPool) Submit(task func()) {
	if p == nil || task == nil {
		return
	}
	if p.closed.Load() {
		return
	}
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// Ignore send-on-closed panic during shutdown races.
		}
	}()
	p.tasks <- task
}

func (p *WorkerPool) Stop() {
	if p == nil {
		return
	}
	p.once.Do(func() {
		p.closed.Store(true)
		close(p.tasks)
		p.wg.Wait()
	})
}
