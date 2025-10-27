# Project Memory Bank

Project Memory Bank is a professional command-line platform for discovering, tracking, and managing local software projects with AI-assisted insights.

## Technology Stack

- **Go 1.21+** – Standalone compiled binary
- **SQLite3** – Embedded, serverless database
- **Cobra** – Enterprise-grade CLI framework
- **Gum** – Interactive TUI components
- **Groq API** – AI analysis with `moonshotai/kimi-k2-instruct`

## Features

### Automatic Detection
- ✅ **Node.js** - package.json, frameworks (React, Next.js, Vue, Angular, Express)
- ✅ **Python** - requirements.txt
- ✅ **Go** - go.mod + version extraction
- ✅ **Rust** - Cargo.toml + edition
- ✅ **Git** - remote, branch, repository information

### Project Management
- Statuses: `active`, `paused`, `archived`, `completed`
- Automatic progress (0-100%)
- Search by name, description, path
- Status filters
- Automatic timestamps

### AI Analysis
- Current state evaluation
- Completion estimation
- Recommended next steps
- Technical blocker identification
- Optimized token consumption

## Installation

```bash
# Clone repository
git clone https://github.com/snowarch/project-memory.git
cd project-memory

# Build
make build

# Install globally (optional)
make install

# Initialize database
pmem init
```

## Configuration

### Groq API Key

```bash
# Environment variable (recommended)
export GROQ_API_KEY='your-api-key-here'

# Or use flag
pmem analyze project-name --api-key='your-api-key'
```

### Database

Default: `~/.local/share/pmem/projects.db`

Customize with `--db` flag:
```bash
pmem scan /path --db=/custom/path/projects.db
```

## Usage

### Available Commands

```bash
# Initialize database
pmem init

# Scan directory for projects
pmem scan /path/to/projects
pmem scan /path/to/projects -v  # verbose

# List projects (interactive TUI with Gum)
pmem list
pmem list --status active
pmem list --status completed --limit 10

# View/change project status
pmem status project-name
pmem status project-name paused
pmem status project-name completed

# AI analysis with Groq
pmem analyze project-name
pmem analyze project-name --api-key='...'
```

### Global Flags

- `-v, --verbose` - Detailed output with DEBUG logs
- `-q, --quiet` - Silence output except errors
- `--db <path>` - Custom path to database
- `-p, --path <path>` - Root path for scan

## Examples

### Typical Workflow

```bash
# 1. Initialize
pmem init

# 2. Scan projects
export GROQ_API_KEY='gsk_...'
pmem scan ~/CascadeProjects -v

# 3. View interactive list
pmem list

# 4. Analyze specific project
pmem analyze my-project

# 5. Update status
pmem status my-project completed
```

### Search and Filters

```bash
# List only active projects
pmem list --status active

# Paused projects
pmem list --status paused

# First 5 projects
pmem list --limit 5
```

## Project Structure

```
project-memory/
├── cmd/pmem/           # Entry point
├── internal/
│   ├── ai/             # Groq API client
│   ├── commands/       # Cobra commands
│   ├── database/       # SQLite setup
│   ├── logger/         # Logging system
│   ├── models/         # Data structures
│   ├── repository/     # Data layer
│   └── scanner/        # Project detection
├── go.mod
├── Makefile
└── README.md
```

## Tests

```bash
# Run all tests
go test -v ./...

# Specific test
go test -v ./internal/scanner
go test -v ./internal/repository

# Coverage
go test -cover ./...
```

**Test Status:** ✅ 23/23 tests passing

## Database

### Schema

- `projects` - Project information
- `technologies` - Detected tech stack
- `project_files` - File metadata
- `todos` - Extracted TODOs
- `ai_analyses` - AI analysis history
- `activity_log` - Change log
- `config` - System configuration

See [`internal/database/schema.sql`](internal/database/schema.sql) for details.

## AI Model

**Current model:** `moonshotai/kimi-k2-instruct`

- Context window: 131,072 tokens
- Max completion: 16,384 tokens
- Optimized for software project analysis

## License

MIT License - See [LICENSE](LICENSE)

## Author

**snowarch** - [GitHub](https://github.com/snowarch)
