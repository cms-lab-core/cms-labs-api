#!/bin/bash
go install gotest.tools/gotestsum@latest
go install github.com/go-critic/go-critic/cmd/gocritic@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/jstemmer/go-junit-report/v2@latest
go install github.com/rubenv/sql-migrate/...@latest
go install golang.org/x/tools/cmd/cover@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/t-yuki/gocover-cobertura@latest
go install github.com/clabernetes/clabernetes/cmd/clabverter@latest
go install github.com/goreleaser/goreleaser/v2@latest
