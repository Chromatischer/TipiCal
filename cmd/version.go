package cmd

import "fmt"

// Version is the current build version.
const Version = "0.1.0-beta.2"

func printVersion() {
	fmt.Println(Version)
}
