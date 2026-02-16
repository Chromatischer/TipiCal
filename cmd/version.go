package cmd

import "fmt"

// Version is the current build version.
const Version = "0.1.0-beta.0"

func printVersion() {
	fmt.Println(Version)
}
