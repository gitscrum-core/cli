# Contributing

Guidelines for contributing to GitScrum CLI.

---

## Getting Started

1. Fork the repository
2. Clone your fork
   ```bash
   git clone https://github.com/YOUR_USERNAME/cli.git
   cd cli
   ```
3. Install dependencies
   ```bash
   go mod download
   ```
4. Build the project
   ```bash
   make build
   ```

---

## Development Workflow

### Building

```bash
make build        # Build binary to ./bin/gitscrum
make install      # Build and install to $GOPATH/bin
```

### Testing

```bash
make test         # Run all tests
make test-cover   # Run tests with coverage
make lint         # Run linter
```

### Running Locally

```bash
./bin/gitscrum --help
./bin/gitscrum tasks
```

---

## Project Structure

```
cli/
├── cmd/gitscrum/       # Main entry point
├── pkg/
│   ├── api/            # HTTP client and API utilities
│   ├── auth/           # Authentication (OAuth, token storage)
│   ├── cmd/            # Command implementations
│   │   ├── auth/
│   │   ├── config/
│   │   ├── projects/
│   │   ├── sprints/
│   │   ├── tasks/
│   │   └── timer/
│   ├── config/         # Configuration management
│   ├── git/            # Git utilities
│   └── output/         # Output formatting (table, JSON)
├── docs/               # Documentation
│   └── examples/       # CI/CD and automation examples
└── install.sh          # Installation script
```

---

## Adding a New Command

1. Create directory under `pkg/cmd/<command>/`
2. Implement the command using Cobra
3. Register in `pkg/cmd/root/root.go`
4. Add tests in `pkg/cmd/<command>/<command>_test.go`
5. Update documentation

**Example structure:**

```go
// pkg/cmd/example/example.go
package example

import (
    "github.com/spf13/cobra"
    "github.com/gitscrum-core/cli/pkg/cmd/factory"
)

func NewCmdExample(f *factory.Factory) *cobra.Command {
    return &cobra.Command{
        Use:   "example",
        Short: "Example command",
        Long:  `Detailed description of the example command.`,
        Example: `  gitscrum example
  gitscrum example --flag value`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
}
```

---

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `make lint` before committing
- Add comments for exported functions
- Keep functions focused and testable

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add sprint burndown command
fix: resolve timer not stopping on branch switch
docs: update README with installation instructions
refactor: simplify API client error handling
test: add unit tests for config package
chore: update dependencies
```

---

## Branch Naming

| Prefix | Use case |
|:-------|:---------|
| `feature/` | New features |
| `bugfix/` | Bug fixes |
| `docs/` | Documentation changes |
| `refactor/` | Code refactoring |
| `test/` | Test additions |

---

## Pull Request Process

1. Create a feature branch
   ```bash
   git checkout -b feature/your-feature
   ```

2. Make changes with tests

3. Run tests locally
   ```bash
   make test
   make lint
   ```

4. Commit with descriptive message
   ```bash
   git commit -m "feat: description of changes"
   ```

5. Push and create PR
   ```bash
   git push origin feature/your-feature
   ```

6. Fill out PR template
   - Describe changes
   - Link related issues
   - Add screenshots if applicable

---

## Testing

### Running Tests

```bash
make test                    # All tests
go test ./pkg/cmd/tasks/...  # Specific package
go test -v ./...             # Verbose output
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests
- Mock HTTP responses for API tests
- Aim for >80% coverage on new code

---

## Reporting Bugs

Include:
- CLI version (`gitscrum --version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Error messages

---

## Feature Requests

Open an issue with:
- Clear description
- Use case and motivation
- Example usage

---

## Questions

- [GitHub Discussions](https://github.com/gitscrum-core/cli/discussions)

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
