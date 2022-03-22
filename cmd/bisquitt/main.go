package main

import (
	"fmt"
	"os"
)

func main() {
	err := Application.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
