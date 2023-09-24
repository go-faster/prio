package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Hello")
	ticker := time.NewTicker(1 * time.Second)
	do := func() {
		fmt.Println("Do")
		cmd := exec.Command("/usr/bin/sleep", "1")
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
	}
	do()
	for range ticker.C {
		do()
	}
}
