# Contributing to webfetch-clean

Thank you for your interest in contributing to webfetch-clean! This document provides guidelines and best practices for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow:

- **Be respectful**: Treat all community members with respect and kindness
- **Be inclusive**: Welcome contributors of all backgrounds and experience levels
- **Be constructive**: Provide helpful feedback and be open to receiving it
- **Be collaborative**: Work together to improve the project
- **Be professional**: Keep discussions focused on technical merits

Unacceptable behavior will not be tolerated. Please report any concerns to the maintainers.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Basic understanding of Go and MCP protocol

### Setting Up Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/webfetch-clean.git
   cd webfetch-clean
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/hegner123/webfetch-clean.git
   ```

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Build the project:**
   ```bash
   go build -o webfetch-clean
   ```

6. **Run tests:**
   ```bash
   go test ./...
   ```

## Development Workflow

### Branching Strategy

We use a simplified Git flow:

- **`main`**: Production-ready code
- **`feature/*`**: New features
- **`fix/*`**: Bug fixes
- **`docs/*`**: Documentation updates
- **`refactor/*`**: Code refactoring

### Creating a Feature Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### Keeping Your Branch Updated

```bash
# Fetch upstream changes
git fetch upstream

# Rebase your branch
git rebase upstream/main
```

## Coding Standards

### Go Style Guide

Follow these Go best practices:

1. **Follow official Go style:**
   - Run `gofmt` on all code
   - Run `go vet` to catch common issues
   - Use `golint` for style suggestions

2. **Code organization:**
   - Keep functions small and focused (single responsibility)
   - Prefer early returns over nested conditionals
   - Avoid deep nesting (max 3-4 levels)
   - Use meaningful variable and function names

3. **Error handling:**
   - Always handle errors explicitly
   - Provide context in error messages
   - Use `fmt.Errorf` with `%w` for error wrapping

4. **Comments:**
   - Add package-level documentation
   - Document all exported functions and types
   - Use complete sentences in comments
   - Explain "why" not "what" in comments

### Example Code Style

```go
// Good: Clear, early return, meaningful names
func ProcessURL(url string, timeout int) (*Result, error) {
    if url == "" {
        return nil, fmt.Errorf("URL cannot be empty")
    }

    html, err := FetchURL(url, timeout)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch URL: %w", err)
    }

    cleaned, err := CleanHTML(html, false, false)
    if err != nil {
        return nil, fmt.Errorf("failed to clean HTML: %w", err)
    }

    return &Result{Content: cleaned}, nil
}

// Bad: Nested, unclear, poor error handling
func ProcessURL(url string, timeout int) (*Result, error) {
    if url != "" {
        html, err := FetchURL(url, timeout)
        if err == nil {
            cleaned, err := CleanHTML(html, false, false)
            if err == nil {
                return &Result{Content: cleaned}, nil
            } else {
                return nil, err
            }
        } else {
            return nil, err
        }
    }
    return nil, fmt.Errorf("error")
}
```

### Performance Guidelines

- Avoid premature optimization
- Profile before optimizing
- Use benchmarks to measure improvements
- Consider memory allocations in hot paths
- Reuse buffers where appropriate

## Testing

### Writing Tests

1. **Test file naming:** `*_test.go`

2. **Test function naming:** `TestFunctionName` or `TestFunctionName_Scenario`

3. **Table-driven tests for multiple cases:**
   ```go
   func TestCleanHTML(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           want    string
           wantErr bool
       }{
           {
               name:    "removes script tags",
               input:   "<html><script>alert('hi')</script><p>content</p></html>",
               want:    "<html><p>content</p></html>",
               wantErr: false,
           },
           // ... more test cases
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got, err := CleanHTML(tt.input, false, false)
               if (err != nil) != tt.wantErr {
                   t.Errorf("CleanHTML() error = %v, wantErr %v", err, tt.wantErr)
                   return
               }
               if got != tt.want {
                   t.Errorf("CleanHTML() = %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

4. **Test coverage:** Aim for at least 70% coverage for new code

5. **Test types:**
   - Unit tests for individual functions
   - Integration tests for end-to-end workflows
   - Edge cases and error conditions

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestFunctionName ./...

# Run benchmarks
go test -bench=. ./...
```

## Commit Guidelines

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring (no feature change or bug fix)
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)

**Examples:**

```
feat(cleaner): add support for removing cookie banners

Implement detection and removal of common cookie consent
banners by checking for class names containing 'cookie',
'consent', and 'gdpr'.

Closes #123
```

```
fix(fetcher): handle redirect loops properly

Add detection for redirect cycles to prevent infinite loops
when fetching URLs. Returns error after 10 redirects.

Fixes #456
```

```
docs: update README with installation troubleshooting

Add troubleshooting section for common installation issues
including dependency conflicts and permission errors.
```

### Commit Best Practices

- Keep commits atomic (one logical change per commit)
- Write clear, descriptive commit messages
- Reference issue numbers where applicable
- Sign commits if possible (`git commit -s`)
- Keep commit history clean (squash WIP commits before PR)

## Pull Request Process

### Before Submitting

1. **Ensure all tests pass:**
   ```bash
   go test ./...
   ```

2. **Run code quality checks:**
   ```bash
   gofmt -w .
   go vet ./...
   golint ./...
   ```

3. **Update documentation** if needed

4. **Add tests** for new functionality

5. **Update CHANGELOG.md** for notable changes

### Creating a Pull Request

1. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR on GitHub** with:
   - Clear, descriptive title
   - Detailed description of changes
   - Reference to related issues
   - Screenshots/examples if applicable
   - Checklist of completed items

3. **PR Template:**
   ```markdown
   ## Description
   Brief description of the changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Related Issues
   Closes #123

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing completed

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Tests pass locally
   - [ ] Documentation updated
   - [ ] No new warnings introduced

   ## Screenshots (if applicable)
   ```

### PR Review Process

1. **Maintainers will review** within 1-3 business days
2. **Address feedback** by pushing additional commits
3. **Keep PR focused** - split large changes into multiple PRs
4. **Be responsive** to review comments
5. **PR will be merged** once approved by a maintainer

### PR Requirements

- All tests must pass
- Code coverage should not decrease
- No merge conflicts with main
- At least one maintainer approval
- All review comments resolved

## Issue Guidelines

### Reporting Bugs

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Run command: `webfetch-clean --url https://example.com`
2. Observe error: ...

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: macOS 14.2
- Go version: 1.23
- webfetch-clean version: 1.0.0

## Additional Context
Error messages, logs, screenshots
```

### Requesting Features

Use the feature request template:

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should it work?

## Alternatives Considered
Other approaches you've thought about

## Additional Context
Examples, mockups, related issues
```

### Issue Best Practices

- Search existing issues before creating new ones
- Use clear, descriptive titles
- Provide complete information
- Use labels appropriately
- Be respectful and constructive
- Update issues with new information

## Documentation

### Documentation Standards

1. **README.md**: Overview, installation, basic usage
2. **Code comments**: Function/type documentation
3. **Examples**: Real-world usage examples
4. **CHANGELOG.md**: Notable changes by version
5. **Architecture docs**: Design decisions and patterns

### Writing Documentation

- Use clear, concise language
- Include code examples
- Keep examples up to date
- Use proper markdown formatting
- Add table of contents for long documents
- Proofread for typos and clarity

## Community

### Getting Help

- **Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Pull Requests**: For code contributions

### Recognition

Contributors are recognized in:
- Pull request acknowledgments
- Release notes
- GitHub contributors page
- Special thanks in documentation

### Maintainer Responsibilities

Maintainers will:
- Review PRs and issues promptly
- Provide constructive feedback
- Maintain project direction and quality
- Foster inclusive community
- Release new versions regularly

## License

By contributing to webfetch-clean, you agree that your contributions will be licensed under the MIT License.

## Questions?

If you have questions about contributing, please:
- Open a discussion on GitHub
- Review existing documentation
- Check closed issues for similar questions

Thank you for contributing to webfetch-clean! 🎉
