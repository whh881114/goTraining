package main

import "fmt"

func f1() int {
	x := 5
	defer func() {
		x++
	}()
	return x // 1.返回值=5 2.x++ 3. return 5
}

func f2() (x int) {
	defer func() {
		x++
	}()
	return 5 // 1.返回值=x(5) 2.x++ 3. return 6
}

func f3() (y int) {
	x := 5
	defer func() {
		x++
	}()
	return x // 1.返回值=y 2.x++ 3. return 5
}

func f4() (x int) {
	defer func(x int) {
		x++
	}(x)
	return 5 // 1.返回值=x(5) 2.x++ 3. return 5
}

func main() {
	fmt.Println(f1())
	fmt.Println(f2())
	fmt.Println(f3())
	fmt.Println(f4())
}
