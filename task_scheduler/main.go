package main

import (
	"fmt"
	"time"
)

func main() {
	t1 := task{
		Id:   1,
		Kind: "CPU",
		Execute: func() (string, error) {
			// Any logic you want to run!
			time.Sleep(100 * time.Millisecond)
			return "Task 1 finished", nil
		},
	}
	t2 := task{
		Id:   2,
		Kind: "CPU",
		Execute: func() (string, error) {
			// Any logic you want to run!
			time.Sleep(140 * time.Millisecond)
			return "Task 2 finished", nil
		},
	}
	p := NewPool(1, 2)
	Start(p)
	Submit(p, t1)
	Submit(p, t2)
	Close(p)
	for res := range p.Resultq {
		fmt.Printf("Task %d: %s (took %v)\n", res.TaskId, res.Outcome, res.Duration)
	}

}
