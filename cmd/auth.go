package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/setup"
)

func handleAuth() {
	if len(os.Args) < 3 {
		printAuthHelp()
		return
	}

	switch os.Args[2] {
	case "add":
		authAdd()
	case "test":
		authTest()
	case "list":
		authList()
	case "help", "--help", "-h":
		printAuthHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown auth command: %s\n\n", os.Args[2])
		printAuthHelp()
		os.Exit(1)
	}
}

func printAuthHelp() {
	fmt.Println("TipiCal Auth Commands")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical auth add     Add a new calendar to your configuration")
	fmt.Println("  tipical auth test    Test connections to configured calendars")
	fmt.Println("  tipical auth list    List all configured calendars")
	fmt.Println()
}

func authAdd() {
	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Println("Run 'tipical setup' first to create a configuration.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Add a new calendar")
	fmt.Println("==================")
	fmt.Println()

	cal := setup.PromptCalendar(reader)
	cfg.Calendars = append(cfg.Calendars, cal)

	// Save updated config
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding config directory: %v\n", err)
		os.Exit(1)
	}

	dir := filepath.Join(configDir, "tipical")
	configPath := filepath.Join(dir, "config.toml")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, buf.Bytes(), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Calendar added successfully!")
	fmt.Printf("Configuration saved to %s\n", configPath)
}

func authTest() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Calendars) == 0 {
		fmt.Println("No calendars configured.")
		fmt.Println("Run 'tipical auth add' to add a calendar.")
		return
	}

	fmt.Println("Testing calendar connections...")
	fmt.Println()

	var targetName string
	if len(os.Args) >= 4 {
		targetName = os.Args[3]
	}

	for i, cal := range cfg.Calendars {
		// If a specific calendar name is provided, only test that one
		if targetName != "" && cal.Name != targetName {
			continue
		}

		fmt.Printf("Testing %q...\n", cal.Name)

		client, err := caldav.NewClient(&cal, i)
		if err != nil {
			fmt.Printf("  ✗ FAILED: %v\n", err)
			fmt.Println()
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		calendars, err := client.DiscoverCalendars(ctx)
		cancel()

		if err != nil {
			fmt.Printf("  ✗ FAILED: %v\n", err)
		} else {
			fmt.Printf("  ✓ SUCCESS: Found %d calendar(s)\n", len(calendars))
			for _, c := range calendars {
				fmt.Printf("    - %s\n", c.Name)
			}
		}
		fmt.Println()
	}

	if targetName != "" {
		found := false
		for _, cal := range cfg.Calendars {
			if cal.Name == targetName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Calendar %q not found in configuration.\n", targetName)
			os.Exit(1)
		}
	}
}

func authList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Calendars) == 0 {
		fmt.Println("No calendars configured.")
		fmt.Println("Run 'tipical auth add' to add a calendar.")
		return
	}

	fmt.Println("Configured calendars:")
	fmt.Println()

	for i, cal := range cfg.Calendars {
		fmt.Printf("%d. %s\n", i+1, cal.Name)
		fmt.Printf("   URL: %s\n", cal.URL)
		fmt.Printf("   User: %s\n", cal.Username)
		if cal.Color != "" {
			fmt.Printf("   Color: %s\n", cal.Color)
		}
		if cal.ReadOnly {
			fmt.Printf("   Read-only: yes\n")
		}
		if cal.PasswordCmd != "" {
			fmt.Printf("   Auth: password command\n")
		} else {
			fmt.Printf("   Auth: password (stored)\n")
		}
		fmt.Println()
	}
}
