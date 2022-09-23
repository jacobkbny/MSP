package main

import (
	"fmt"
	"strings"
)

func main() {

	abc := (3.0 / 13.0) * 100
	fmt.Println(abc)
	temp := fmt.Sprint(abc)
	target := strings.Split(temp, ".")
	fmt.Println(target[0])

}
