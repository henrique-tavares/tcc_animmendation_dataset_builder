package workerpool

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

type WorkerPool[In, Out any] struct {
	wg       sync.WaitGroup
	size     int
	jobs     <-chan In
	results  chan<- Out
	callback func(In, func(string)) Out
}

func New[In, Out any](size int, jobs <-chan In, results chan<- Out, callback func(In, func(string)) Out) *WorkerPool[In, Out] {
	return &WorkerPool[In, Out]{
		wg:       sync.WaitGroup{},
		size:     size,
		jobs:     jobs,
		results:  results,
		callback: callback,
	}
}

func (self *WorkerPool[In, Out]) worker(id int) {
	log := logStep(id, self.size)
	for job := range self.jobs {
		output := self.callback(job, log)
		self.results <- output
	}
	log("finished!")
	self.wg.Done()
}

func logStep(workerId int, size int) func(string) {
	digits := int(math.Log10(float64(size))) + 1

	return func(msg string) {
		moveUp := fmt.Sprintf("\033[%vA\033[K", size-workerId)
		moveDown := fmt.Sprintf("\033[%vB\r", size-workerId)
		fmt.Printf("%v[Worker #%v] %v%v", moveUp, fmt.Sprintf("%0*d", digits, workerId+1), msg, moveDown)
	}
}

func (self *WorkerPool[In, Out]) Run() {
	self.wg.Add(self.size)

	fmt.Print(strings.Repeat("\n", self.size+1))
	for i := 0; i < self.size; i++ {
		// fmt.Printf("Worker #%v started\n", i+1)
		go self.worker(i)
	}
	self.wg.Wait()
	fmt.Println()
	close(self.results)
}
