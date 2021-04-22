 package main

import (
	"fmt"
)

func warn(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
