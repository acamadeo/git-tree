package main

import "fmt"

import "github.com/zyedidia/generic/queue"

func main() {
	q := queue.New[int]()
	q.Enqueue(1)

	fmt.Println(q.Dequeue())
}
