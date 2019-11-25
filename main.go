package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	done := make(chan struct{})

	plus1 := mapIntChan(done, func(x int) int { return x + 1 }, intChan(done, 10))
	times2 := mapIntChan(done, func(x int) int { return x * 2 }, intChan(done, 10))
	sq := mapIntChan(done, func(x int) int { return x * x }, intChan(done, 10))

	fmt.Println(intChanToSlice(done, merge(done, plus1, times2, sq)))

	<-time.After(time.Second)
}

func intChan(done <-chan struct{}, n int) <-chan int {
	out := make(chan int)

	go func() {
		defer func() {
			log.Println("clossing intChan", n)
			close(out)
		}()

		for i := 0; i < n; i++ {
			select {
			case out <- i:
			case <-done:
				return
			}
		}
	}()

	return out
}

func mapIntChan(done <-chan struct{}, mapper func(int) int, in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer func() {
			log.Println("clossing mapIntChan")
			close(out)
		}()

		for x := range in {
			select {
			case out <- mapper(x):
			case <-done:
				return
			}
		}
	}()

	return out
}

func intChanToSlice(done chan<- struct{}, in <-chan int) []int {
	out := []int{}

	for x := range in {
		out = append(out, x)
		if len(out) == 20 {
			break
		}
	}
	close(done)

	return out
}

func merge(done <-chan struct{}, ins ...<-chan int) <-chan int {
	out := make(chan int)

	wg := sync.WaitGroup{}

	forward := func(in <-chan int) {
		defer wg.Done()

		for x := range in {
			select {
			case out <- x:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(ins))
	for _, in := range ins {
		go forward(in)
	}

	go func() {
		wg.Wait()
		log.Println("clossing merge")
		close(out)
	}()

	return out
}
