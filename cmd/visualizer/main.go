package main

import (
	"fmt"
	"os"

	"lem-in/pkg/visualizer"
)

func main() {
	err := visualizer.RunVisualizer(os.Stdin)
	if err != nil {
		fmt.Println("Visualizer error:", err)
		os.Exit(1)
	}
}