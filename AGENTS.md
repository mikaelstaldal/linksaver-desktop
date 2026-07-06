### Project Guidelines

#### Build/Configuration Instructions
The project is written in Go and uses GTK4. 
- **Requirements**: Go 1.25 or later.
- **Build**: Use standard Go build commands:
  ```bash
  go build -o mylinks-desktop .
  ```
- **Dependencies**: Dependencies are managed via Go modules. Run `go mod tidy` to ensure all dependencies are resolved.

#### Testing Information
- **Running Tests**: Use the standard Go test tool:
  ```bash
  go test ./...
  ```
- **Adding Tests**: Create files with the `_test.go` suffix. Since this is a GUI application, unit tests should ideally focus on logic separated from the UI.
- **Example Test**:
  ```go
  package main

  import "testing"

  func TestExample(t *testing.T) {
      // Your test logic here
  }
  ```

#### Additional Development Information
- **GUI Library**: This project uses GTK4
- **Code Style**: Follow standard Go formatting (`gofmt`).
