//author: richard
package main

import "fmt"

func main() {
	var first, last string
	_, err := fmt.Scanln(&first, &last)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(first, last)
}
