package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Needed for random ClientID generation.
	rand.Seed(time.Now().UnixNano())

	err := Application.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
