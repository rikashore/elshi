package main

import (
	"elshi/internal/vm"
	"fmt"
	"os"
)

func main() {
	byts, err := os.ReadFile("examples/rogue.obj")

	if err != nil {
		fmt.Printf("%e\n", err)
	}

	v, err := vm.NewVM(byts)

	if err != nil {
		fmt.Printf("%e\n", err)
		os.Exit(1)
	}

	v.Execute()

}
