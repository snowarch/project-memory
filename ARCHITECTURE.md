# Project Memory Bank - Architecture

## Overview

Professional CLI tool for tracking, analyzing, and managing development projects with AI-powered insights.

## Technical Stack

- Language: Go 1.25+
- Database: SQLite3
- TUI: Gum (Charm)
- AI: Groq API (llama-3.3-70b-versatile)
- CLI Framework: Cobra

## Core Components

### 1. Scanner
Location: `internal/scanner/scanner.go`

Responsibilities:
- Scan directories for project indicators
- Extract basic project metadata
- Detect technologies from package managers
- Extract Git information

Indicators:
- package.json (Node.js)
- requirements.txt (Python)
- go.mod (Go)
- Cargo.toml (Rust)
- .git directory
- README.md

### 2. Database Layer
Location: `internal/database/`

Schema:
- projects: Core project information
- technologies: Detected tech stack per project
- project_files: File metadata
- todos: Extracted TODO items
- ai_analyses: AI analysis history
- activity_log: Change tracking
- config: System configuration

### 3. Repository Pattern
Location: `internal/repository/`

Repositories:
- ProjectRepository: CRUD operations for projects
- TechnologyRepository: Tech stack management
- AnalysisRepository: AI analysis storage

### 4. AI Integration
Location: `internal/ai/groq.go`

Capabilities:
- Project status analysis
- Progress estimation
- Next steps recommendation
- TODO summarization

API: Groq (OpenAI-compatible)
Model: llama-3.3-70b-versatile

### 5. Commands
Location: `internal/commands/`

Available commands:
- init: Initialize database
- scan: Scan directories for projects
- list: Interactive project browser (Gum)
- analyze: AI-powered project analysis
- status: View/update project status

## Data Flow

### Scan Flow
```
User → scan command → Scanner
  → Detect projects
  → Extract metadata
  → Detect technologies
  → Save to database
```

### List Flow
```
User → list command
  → Query database
  → Format for Gum
  → Interactive selection
  → Display details
```

### Analysis Flow
```
User → analyze command
  → Fetch project data
  → Read README/TODO
  → Call Groq API
  → Parse response
  → Save analysis
  → Display results
```

## Database Schema

### projects
- id: TEXT PRIMARY KEY (SHA256 hash of path)
- name: TEXT (directory name)
- path: TEXT UNIQUE (absolute path)
- description: TEXT (from README)
- status: TEXT (active|paused|archived|completed)
- progress: INTEGER (0-100)
- created_at: INTEGER (Unix timestamp)
- updated_at: INTEGER (Unix timestamp)
- last_scanned_at: INTEGER (Unix timestamp)
- is_git_repo: BOOLEAN
- git_remote: TEXT
- git_branch: TEXT
- notes: TEXT

### technologies
- id: INTEGER PRIMARY KEY AUTOINCREMENT
- project_id: TEXT (FK to projects)
- type: TEXT (runtime|dependency|framework)
- name: TEXT
- version: TEXT
- detected_from: TEXT (package.json, requirements.txt, etc.)

### ai_analyses
- id: INTEGER PRIMARY KEY AUTOINCREMENT
- project_id: TEXT (FK to projects)
- analysis_type: TEXT (project_status, todo_summary, etc.)
- result: TEXT (analysis content)
- model: TEXT (AI model used)
- tokens_used: INTEGER
- analyzed_at: INTEGER (Unix timestamp)

## Technology Detection

### Node.js Projects
Source: package.json
Extracts: dependencies, devDependencies
Runtime: Node.js

### Python Projects
Source: requirements.txt
Extracts: package==version
Runtime: Python

### Go Projects
Source: go.mod
Extracts: module path
Runtime: Go

### Git Projects
Source: .git/config, .git/HEAD
Extracts: remote URL, current branch

## Gum Integration

Interactive elements:
- `gum choose`: Project selection
- `gum pager`: Project details viewer
- `gum input`: Text input (future)
- `gum confirm`: Confirmations (future)

## AI Analysis

### Project Status Analysis
Input:
- Project name
- Description
- Technology list
- README excerpt

Output:
- Current state assessment
- Estimated completion percentage
- Key next steps
- Technical concerns

### Token Optimization
- README truncated to 4000 chars
- Prompt templates optimized
- Temperature: 0.3 (focused responses)
- Max tokens: 2000

## Build System

Makefile targets:
- build: Compile binary
- install: Install to /usr/local/bin
- clean: Remove artifacts
- test: Run tests
- deps: Download dependencies

## Configuration

Environment variables:
- GROQ_API_KEY: Groq API key for AI analysis

Database location:
- Default: ~/.local/share/pmem/projects.db
- Override: --db flag

## Security

- API keys via environment variables only
- Database stored in user directory
- No network access except Groq API
- Local-first architecture

## Future Enhancements

Potential additions:
1. TODO extraction from source files
2. Recent activity tracking
3. Dependency vulnerability scanning
4. Custom project templates
5. Export to JSON/CSV
6. Web dashboard (optional)
7. Team collaboration features
8. Git integration (commits, branches)
9. CI/CD status integration

## Performance Considerations

- SQLite for low overhead
- Lazy loading of project details
- Cached technology detection
- Batch operations for scanning
- Index optimization on common queries

## Error Handling

Strategy:
- Graceful degradation
- Clear error messages
- Fallback to defaults
- No data loss on failures

## Testing Strategy

Test coverage:
- Scanner unit tests
- Repository integration tests
- Command E2E tests
- Database migration tests

## Deployment

Binary distribution:
1. Single binary (Go)
2. No external dependencies except Gum
3. Database auto-initialized
4. Cross-platform compatible
