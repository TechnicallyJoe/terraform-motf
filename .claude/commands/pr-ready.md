Run the full pre-merge checklist to verify this branch is ready for PR:

1. go build -o motf ./cmd/motf
2. go test ./...
3. golangci-lint run
4. go vet ./...

Run all four steps. Report a pass/fail summary for each. If any step fails, show the errors and suggest fixes.
