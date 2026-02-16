package setup

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"golang.org/x/term"
)

// RunWizard runs the interactive setup wizard and writes config.toml.
func RunWizard() error {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.DefaultConfig()

	fmt.Println("TipiCal setup wizard")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("This will create a configuration file for TipiCal.")
	fmt.Println("Press Enter to accept the default value shown in [brackets].")
	fmt.Println()

	// --- General settings ---
	fmt.Println("-- General Settings --")
	fmt.Println()

	cfg.General.DefaultView = config.ViewType(Prompt(reader,
		"Default view (month, week, threeday, day, agenda, stacked)",
		string(cfg.General.DefaultView)))

	cfg.General.WeekStart = Prompt(reader,
		"Week starts on (monday, sunday)",
		cfg.General.WeekStart)

	cfg.General.TimeFormat = Prompt(reader,
		"Time format (24h, 12h)",
		cfg.General.TimeFormat)

	fmt.Println()

	// --- Calendars ---
	fmt.Println("-- Calendars --")
	fmt.Println()

	for {
		cal := PromptCalendar(reader)
		cfg.Calendars = append(cfg.Calendars, cal)

		fmt.Println()
		if !PromptYesNo(reader, "Add another calendar?", false) {
			break
		}
		fmt.Println()
	}

	fmt.Println()

	// --- Sync settings ---
	fmt.Println("-- Sync Settings --")
	fmt.Println()

	intervalStr := Prompt(reader, "Sync interval in minutes", fmt.Sprintf("%d", cfg.Sync.IntervalMinutes))
	fmt.Sscanf(intervalStr, "%d", &cfg.Sync.IntervalMinutes)
	if cfg.Sync.IntervalMinutes < 1 {
		cfg.Sync.IntervalMinutes = 5
	}

	fmt.Println()

	// --- Write config ---
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("finding config directory: %w", err)
	}

	dir := filepath.Join(configDir, "tipical")
	configPath := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(configPath); err == nil {
		if !PromptYesNo(reader, fmt.Sprintf("%s already exists. Overwrite?", configPath), false) {
			fmt.Println("Aborted. No changes were made.")
			return nil
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(configPath, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Println("Configuration saved to", configPath)
	fmt.Println("Run tipical to start the app.")
	return nil
}

// PromptCalendar prompts for calendar configuration interactively.
func PromptCalendar(reader *bufio.Reader) config.CalendarConfig {
	var cal config.CalendarConfig

	cal.Name = Prompt(reader, "Calendar name", "")
	cal.URL = Prompt(reader, "CalDAV URL", "")
	cal.Username = Prompt(reader, "Username", "")

	useCmd := PromptYesNo(reader, "Use a command for password (e.g. pass, secret-tool)?", false)
	if useCmd {
		cal.PasswordCmd = Prompt(reader, "Password command", "")
	} else {
		cal.Password = PromptPassword(reader)
	}

	cal.Color = Prompt(reader, "Color (hex, e.g. #FF5733, or empty for default)", "")
	cal.ReadOnly = PromptYesNo(reader, "Read-only?", false)

	if PromptYesNo(reader, "Test connection now?", true) {
		testConnection(&cal)
	}

	return cal
}

func testConnection(cal *config.CalendarConfig) {
	fmt.Print("  Testing connection... ")
	client, err := caldav.NewClient(cal, 0)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	calendars, err := client.DiscoverCalendars(ctx)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		return
	}

	fmt.Printf("OK (%d calendar(s) found)\n", len(calendars))
	for _, c := range calendars {
		fmt.Printf("    - %s\n", c.Name)
	}
}

// Prompt prompts the user for input with an optional default value.
func Prompt(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}

	line, _ := reader.ReadString('\n')
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultVal
	}
	return val
}

// PromptYesNo prompts the user for a yes/no answer.
func PromptYesNo(reader *bufio.Reader, label string, defaultYes bool) bool {
	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}
	fmt.Printf("  %s [%s]: ", label, hint)

	line, _ := reader.ReadString('\n')
	val := strings.TrimSpace(strings.ToLower(line))
	if val == "" {
		return defaultYes
	}
	return val == "y" || val == "yes"
}

// PromptPassword prompts the user for a password (hidden input if terminal supports it).
func PromptPassword(reader *bufio.Reader) string {
	fmt.Print("  Password: ")
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		// Flush any buffered data from the reader before reading directly from fd.
		// Discard buffered bytes so the fd position and reader are in sync.
		if reader.Buffered() > 0 {
			reader.Discard(reader.Buffered())
		}
		pw, err := term.ReadPassword(fd)
		fmt.Println() // newline after hidden input
		if err == nil {
			fmt.Printf("  (password set, %d characters)\n", len(pw))
			return string(pw)
		}
		fmt.Printf("  (hidden input failed: %v, falling back to visible input)\n", err)
		fmt.Print("  Password (visible): ")
	}
	line, _ := reader.ReadString('\n')
	pw := strings.TrimSpace(line)
	fmt.Printf("  (password set, %d characters)\n", len(pw))
	return pw
}
