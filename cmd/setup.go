package cmd

import (
	"fmt"
	"os"

	"github.com/tipical/tipical/setup"
)

func runSetup() {
	if err := setup.RunWizard(); err != nil {
		fmt.Fprintf(os.Stderr, "Setup error: %v\n", err)
		os.Exit(1)
	}
}
