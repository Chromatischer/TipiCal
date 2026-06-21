package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func handleCalendars(args []string) {
	fs := flag.NewFlagSet("calendars", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "output JSON")
	fs.Parse(args)

	_, store, _, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		type cal struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Source   string `json:"source,omitempty"`
			Color    string `json:"color,omitempty"`
			ReadOnly bool   `json:"read_only"`
		}
		out := make([]cal, 0, len(store.Calendars))
		for _, c := range store.Calendars {
			out = append(out, cal{ID: c.ID, Name: c.Name, Source: c.Source, Color: c.Color, ReadOnly: c.ReadOnly})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return
	}

	if len(store.Calendars) == 0 {
		fmt.Println("No calendars synced.")
		fmt.Println("Run 'tipical auth add' to add one, or 'tipical auth test' to check connectivity.")
		return
	}

	fmt.Println("Synced calendars:")
	fmt.Println()
	for _, c := range store.Calendars {
		fmt.Printf("%d. %s\n", c.ID, c.Name)
		if c.Source != "" {
			fmt.Printf("   Account: %s\n", c.Source)
		}
		if c.Color != "" {
			fmt.Printf("   Color: %s\n", c.Color)
		}
		if c.ReadOnly {
			fmt.Printf("   Read-only: yes\n")
		}
	}
}
