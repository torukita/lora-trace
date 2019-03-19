package main

import (
	"fmt"
	"github.com/torukita/lora-trace/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

