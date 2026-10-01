.PHONY: clean critic security lint test generate build run

APP_NAME = pnetlabaddon
CUR_DIR = $(CURDIR)
BUILD_DIR = $(CURDIR)/pnetlabaddon/usr/local/bin
SQL_MIGRATE_VERSION = v1.8.1

clean:
	rm -rf ./build

critic:
	gocritic check -enableAll ./...

security:
	gosec ./...

lint:
	golangci-lint run ./... --timeout=10m

test:
	go test -v -timeout 30s -coverprofile=cover.out -cover ./...
	go tool cover -func=cover.out

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install github.com/rubenv/sql-migrate/...@$(SQL_MIGRATE_VERSION)
	cp $(GOPATH)/bin/sql-migrate $(BUILD_DIR)/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME) main.go
	mkdir -p $(CURDIR)/pnetlabaddon/etc/pnetlabaddon/platform/migrations
	cp -R $(CURDIR)/platform/migrations $(CURDIR)/pnetlabaddon/etc/pnetlabaddon/platform
	cp -R $(CURDIR)/migrate.sh $(CURDIR)/pnetlabaddon/etc/pnetlabaddon

run: generate build
	$(BUILD_DIR)/$(APP_NAME)

migrate.up:
	./migrate.sh up

migrate.down:
	./migrate.sh d

migrate.status:
	./migrate.sh s

generate:
	go generate

crud:
	cd ../gen && go run . -model $(MODEL) -root "github.com/cms-lab-core/cms-labs-api/pnetlabaddon" -directory $(CUR_DIR) && cd $(CUR_DIR)
