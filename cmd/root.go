package cmd

import (
	"fmt"
	"os"
)

// Execute runs the main command dispatcher
func Execute() {
	if len(os.Args) < 2 {
		// No subcommand, run the main app
		runApp(false)
		return
	}

	switch os.Args[1] {
	case "demo", "--demo":
		runApp(true)
	case "print":
		runPrint(os.Args[2:])
	case "setup":
		runSetup()
	case "auth":
		handleAuth()
	case "events", "event":
		handleEvents(os.Args[2:])
	case "calendars", "calendar", "cals":
		handleCalendars(os.Args[2:])
	case "mcp":
		runMCP(os.Args[2:])
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("TipiCal - Terminal Calendar")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical              Start the calendar app")
	fmt.Println("  tipical --demo        Start with demo data")
	fmt.Println("  tipical print         Print agenda to console")
	fmt.Println("  tipical setup        Run the setup wizard")
	fmt.Println("  tipical auth add     Add a new calendar")
	fmt.Println("  tipical auth test    Test calendar connections")
	fmt.Println("  tipical auth list    List configured calendars")
	fmt.Println("  tipical calendars    List synced calendars (with ids)")
	fmt.Println("  tipical events list  List events in a date range")
	fmt.Println("  tipical events add   Create an event")
	fmt.Println("  tipical events show  Show an event by UID")
	fmt.Println("  tipical events delete Delete an event by UID")
	fmt.Println("  tipical mcp          Run the MCP server (stdio)")
	fmt.Println("  tipical version      Show version")
	fmt.Println("  tipical help         Show this help message")
	fmt.Println()
}
