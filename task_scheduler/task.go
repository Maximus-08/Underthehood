package main

import "time"

type task struct {
	Id      int
	Kind    string
	Execute func() (string, error)
}

type result struct {
	TaskId   int
	Outcome  string
	Err      error
	Duration time.Duration
}
