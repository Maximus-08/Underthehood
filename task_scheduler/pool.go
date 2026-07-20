package main

import (
	"sync"
	"time"
)

type pool struct {
	Taskq      chan task
	Resultq    chan result
	numWorkers int
	wg         sync.WaitGroup
}

func NewPool(numWorkers int, maxCapacity int) *pool {
	p := pool{}
	p.numWorkers = numWorkers
	p.Taskq = make(chan task, maxCapacity)
	p.Resultq = make(chan result, maxCapacity)
	return &p
}

func Start(p *pool) error {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go workerRoutine(p)
	}
	return nil
}

func workerRoutine(p *pool) {
	for currtask := range p.Taskq {
		start := time.Now()
		res, err := currtask.Execute()
		elapsed := time.Since(start)
		p.Resultq <- result{
			TaskId:   currtask.Id,
			Outcome:  res,
			Err:      err,
			Duration: elapsed,
		}
	}
	p.wg.Done()
}

func Submit(p *pool, t task) {
	p.Taskq <- t
}

func Close(p *pool) {
	close(p.Taskq)
	p.wg.Wait()
	close(p.Resultq)
}
