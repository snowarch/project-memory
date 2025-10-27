# Contributing to Project Memory Bank

Thank you for your interest in contributing to Project Memory Bank.

## Development Process

### 1. Fork and Clone

```bash
# Fork on GitHub, then:
git clone https://github.com/your-username/project-memory.git
cd project-memory
```

### 2. Setup

```bash
# Install dependencies
go mod download

# Build
make build

# Run tests
go test -v ./...
```

### 3. Create Branch

```bash
git checkout -b feature/your-feature
# or
git checkout -b fix/your-bug-fix
```

### 4. Development

- Write clean and tested code
- Follow Go conventions (gofmt, golint)
- Add tests for new features
- Update documentation if necessary

### 5. Tests

```bash
# Unit tests
go test -v ./...

# Tests with coverage
go test -cover ./...

# Race detector
go test -race ./...

# Lint
golangci-lint run
```

### 6. Commit

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for Java projects
fix: correct version detection in Go
docs: update README with examples
test: add tests for scanner
refactor: simplify repository logic
```

### 7. Push and PR

```bash
git push origin feature/your-feature
```

Create Pull Request on GitHub with:
- Descriptive title
- Clear description of changes
- Screenshots if applicable
- References to related issues

## Code Guidelines

### Structure

```go
// Comments for public functions
func PublicFunction(arg string) error {
    // Validation first
    if arg == "" {
        return fmt.Errorf("arg cannot be empty")
    }
    
    // Logic
    result := processArg(arg)
    
    return nil
}
```

### Testing

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "test",
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
            }
            
            if result != tt.expected {
                t.Errorf("expected: %v, got: %v", tt.expected, result)
            }
        })
    }
}
```

### Logging

```go
// Use logger instead of fmt
logger.Debug("Processing project: %s", projectName)
logger.Info("Scan complete: %d projects", count)
logger.Warn("Failed to detect tech: %v", err)
logger.Error("Critical error: %v", err)
```

## Contribution Areas

### High Priority
- [ ] Support for more languages (Java, C++, C#)
- [ ] TODO extraction from source code
- [ ] AI analysis caching
- [ ] Rate limiting for API calls

### Medium Priority
- [ ] `search` command for advanced search
- [ ] Export to JSON/CSV
- [ ] Git commit history analysis
- [ ] Dependency vulnerability scanning

### Low Priority
- [ ] Web dashboard (optional)
- [ ] Plugin system
- [ ] Custom templates
- [ ] Multi-language support in UI

## Bug Reporting

Use [GitHub Issues](https://github.com/snowarch/project-memory/issues) with:

- **Clear title**: "Error scanning Rust projects"
- **Description**: What you expected vs what happened
- **Steps to reproduce**: Exact steps
- **Environment**: OS, Go version, pmem version
- **Logs**: Output with `--verbose` if applicable

## Questions

For general questions, use [GitHub Discussions](https://github.com/snowarch/project-memory/discussions).

## Code of Conduct

- Be respectful and constructive
- Accept feedback openly
- Focus on what's best for the project
- Help other contributors

## License

By contributing, you agree that your code is licensed under MIT License.
