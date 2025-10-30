package worker

import (
	"context"
	"sync"
)

// WorkerPool manages a pool of workers for parallel transaction execution
type WorkerPool interface {
	// Start initializes and starts all workers
	Start(ctx context.Context)

	// Stop gracefully stops all workers
	Stop()

	// SubmitTask submits a task for execution
	SubmitTask(task *Task)

	// GetResults returns a channel for receiving results
	GetResults() <-chan *Result

	// Wait waits for all workers to finish
	Wait()
}

// pool implements WorkerPool
type pool struct {
	workerCount int
	workers     []*Worker
	taskChan    chan *Task
	resultChan  chan *Result
	wg          *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, taskBufferSize int) WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &pool{
		workerCount: workerCount,
		workers:     make([]*Worker, workerCount),
		taskChan:    make(chan *Task, taskBufferSize),
		resultChan:  make(chan *Result, taskBufferSize),
		wg:          &sync.WaitGroup{},
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start initializes and starts all workers
func (p *pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(i, p.taskChan, p.resultChan, p.wg)
		p.workers[i] = worker
		worker.Start(ctx)
	}
}

// Stop gracefully stops all workers
func (p *pool) Stop() {
	// Close task channel to signal workers
	close(p.taskChan)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close result channel
	close(p.resultChan)

	// Cancel context
	p.cancel()
}

// SubmitTask submits a task for execution
func (p *pool) SubmitTask(task *Task) {
	p.taskChan <- task
}

// GetResults returns a channel for receiving results
func (p *pool) GetResults() <-chan *Result {
	return p.resultChan
}

// Wait waits for all workers to finish
func (p *pool) Wait() {
	p.wg.Wait()
}

// GetWorkerCount returns the number of workers in the pool
func (p *pool) GetWorkerCount() int {
	return p.workerCount
}

// IsRunning checks if the pool is running
func (p *pool) IsRunning() bool {
	select {
	case <-p.ctx.Done():
		return false
	default:
		return true
	}
}
