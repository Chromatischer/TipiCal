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
		os.Exit(runPrint(os.Args[2:]))
	case "agenda":
		os.Exit(runAgenda(os.Args[2:]))
	case "calendars":
		os.Exit(runCalendars(os.Args[2:]))
	case "event":
		os.Exit(runEvent(os.Args[2:]))
	case "mcp":
		os.Exit(runMCP(os.Args[2:]))
	case "setup":
		runSetup()
	case "auth":
		handleAuth()
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
	fmt.Println("  tipical --demo       Start with demo data")
	fmt.Println("  tipical print        Print agenda to console")
	fmt.Println("  tipical agenda       Quick agenda view, optionally scoped to one calendar")
	fmt.Println("  tipical calendars    List discovered calendars")
	fmt.Println("  tipical event        Create, update, move, and delete events")
	fmt.Println("  tipical mcp          Run an MCP server over stdio for AI clients")
	fmt.Println("  tipical setup        Run the setup wizard")
	fmt.Println("  tipical auth add     Add a new calendar")
	fmt.Println("  tipical auth test    Test calendar connections")
	fmt.Println("  tipical auth list    List configured calendars")
	fmt.Println("  tipical version      Show version")
	fmt.Println("  tipical help         Show this help message")
	fmt.Println()
}
