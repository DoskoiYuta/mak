package main

import (
	"fmt"
	"os"

	"github.com/doskoiyuta/mak/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "エラー: "+err.Error())
		os.Exit(1)
	}
}
