# Speech Practice Application

A Go web application that helps users improve speech clarity and fluency through structured daily practice exercises. It runs entirely in the browser via server-side rendered HTML — no JavaScript framework required.

## Features

- Exercise library with mouth exercises, tongue twisters, diction strategies, and pacing techniques
- Targeted practice for specific sounds (/s/, /z/, /r/, /l/, /th/, /sh/, /ch/, /j/, /k/, /g/)
- In-browser exercise timer with audible completion signal
- Practice session tracking with streak and progress metrics
- Personalized exercise recommendations based on completion history
- Accessible UI with high-contrast mode, adjustable text size, and keyboard navigation

## Requirements

- Go 1.21 or later

## Getting Started

```bash
# Clone the repo
git clone <repo-url>
cd speech-practice-app

# Run the server
go run cmd/main.go
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

To use a different port:

```bash
PORT=9090 go run cmd/main.go
```

## Running Tests

```bash
go test ./...
```

## Project Structure

```
cmd/            # Application entry point
internal/
  domain/       # Core business entities and types
  infrastructure/ # Repository implementations (in-memory)
  service/      # Business logic layer
  handler/      # HTTP handlers and routing
  repository/   # Repository interfaces
templates/      # Go html/template files
static/         # CSS and static assets
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a detailed developer guide.

## Tech Stack

- Go standard library (`net/http`, `html/template`)
- CSS for styling — no frontend framework
- In-memory storage (file-based persistence planned)

## License

GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
