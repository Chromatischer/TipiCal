# TipiCal

A fast, beautiful calendar interface for your terminal.

![TipiCal UI](docs/images/main.png)

## Features

- **CalDAV Sync** — Connect to any CalDAV server (Nextcloud, Google Calendar, iCloud, etc.)
- **Multiple Views** — Day, 3-day, week, month, agenda, and stacked views
- **Rich Event Editor** — Create and edit events with recurrence, reminders, and more
- **Fast Navigation** — Vim-style keybindings and intuitive controls
- **Themeable** — Customizable colors and styling via TOML config
- **Offline Support** — Local cache with background sync

![Calendar Views](docs/images/views.png)

## Installation

```bash
go install github.com/tipical/tipical@latest
```

Or build from source:

```bash
git clone https://github.com/tipical/tipical
cd tipical
go build -o tipical
```

## Quick Start

Run the setup wizard on first launch:

```bash
tipical setup
```

Or start immediately with demo data:

```bash
tipical
```

Configuration is stored in `~/.config/tipical/config.toml`.

### Managing Calendars

After initial setup, you can manage calendars using the `auth` commands:

```bash
tipical auth add      # Add a new calendar
tipical auth list     # List configured calendars
tipical auth test     # Test calendar connections
```

![Event Editor](docs/images/editor.png)

## Keybindings

| Key | Action |
|-----|--------|
| `h` `j` `k` `l` | Navigate |
| `n` | New event |
| `Enter` | View/edit event |
| `d` `w` `m` | Day/week/month view |
| `t` | Go to today |
| `/` | Search |
| `q` | Quit |

Press `?` in the app to see all keybindings.

## CalDAV Configuration

TipiCal supports any CalDAV server. Example config:

```toml
[[calendars]]
name = "Personal"
url = "https://nextcloud.example.com/remote.php/dav"
username = "user"
password = "app-password"
```

## License

MIT
