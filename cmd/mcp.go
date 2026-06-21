package cmd

import (
	"fmt"
	"os"

	"github.com/tipical/tipical/mcpserver"
)

func runMCP(args []string) {
	for _, arg := range args {
		switch arg {
		case "--help", "-h", "help":
			printMCPHelp()
			return
		}
	}

	// Sync at startup so the server starts with fresh data; reads then serve
	// from the in-memory store. stdout is reserved for the JSON-RPC protocol,
	// so loadStore's warnings go to stderr only.
	cfg, store, syncMgr, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := mcpserver.Serve(cfg, store, syncMgr); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func printMCPHelp() {
	fmt.Println("TipiCal MCP - Model Context Protocol server")
	fmt.Println()
	fmt.Println("Runs an MCP server over stdio that exposes your calendars to LLM agents.")
	fmt.Println("It is meant to be launched by an MCP client (e.g. Claude Desktop or")
	fmt.Println("Claude Code) rather than run interactively.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical mcp")
	fmt.Println()
	fmt.Println("Example client config entry:")
	fmt.Println()
	fmt.Println("  {")
	fmt.Println("    \"mcpServers\": {")
	fmt.Println("      \"tipical\": { \"command\": \"tipical\", \"args\": [\"mcp\"] }")
	fmt.Println("    }")
	fmt.Println("  }")
	fmt.Println()
}
