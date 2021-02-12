package main

import (
	"fmt"
	"sync"
)

var onlyOnce sync.Once

func f1(a int) { fmt.Println(a) }

func closer(x int) func() {
	return func() {
		f1(x)
	}
}

func main() {
	f := closer(10) // 利用闭包实现
	onlyOnce.Do(f)
}
