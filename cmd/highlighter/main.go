package main

import (
	"fmt"
	"os"

	"github.com/dextryz/highlighter"
)

func main() {
	err := highlighter.Main()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
