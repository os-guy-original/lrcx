# Contributing to srt2lrc

## Code Style

- Format all code with `gofmt` before committing
- Keep functions small and focused — one responsibility per function
- No external dependencies inside `internal/` packages
- Errors must be returned, not swallowed

## Submitting a PR

1. Fork the repo and create a feature branch
2. Write or update tests for your change
3. Run `make test` and `make lint` — both must pass
4. Open a PR with a clear description of what and why

**PR checklist:**
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` clean
- [ ] Test coverage ≥ 80% for changed packages
- [ ] No new external dependencies without prior discussion

## AI-Generated Contributions

AI-generated code is welcome, but subject to stricter review:

- **Disclose it** — note in your PR description that the code was AI-generated (fully or partially)
- **You own it** — you are responsible for understanding and verifying every line you submit
- **Logic audit** — a maintainer will audit the logic manually, not just the diff
- **No hallucinated APIs** — all stdlib and package calls must be verified against actual Go docs
- **Test coverage required** — AI-generated code must include tests; coverage ≥ 80%
- **No AI-generated tests for AI-generated code without human review** — at least one test must be written or verified by a human

Undisclosed AI contributions that are later identified will be closed without merge.

## Adding a New Tag Handler

Tag handlers live in `internal/tags/tags.go` in the `registry` map:

```go
var registry = map[string]Handler{
    "i": func(s string) string { return "_" + s + "_" },
    // Add new handlers here, e.g.:
    // "b": func(s string) string { return "*" + s + "*" },
}
```

Each handler receives the inner text of the tag and returns the transformed string.

## License

By contributing, you agree your code will be licensed under the MIT License.
