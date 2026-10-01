.PHONY: clean security pre_commit test generate dev \
	dev-deps-up dev-cluster-up dev-up dev-backend dev-seed dev-clabgate \
	dev-front \
	dev-cluster-down dev-down

DEV_COMPOSE = docker compose -f backend/docker-compose-dev.yml
DEV_KUBECONFIG = $(CURDIR)/.local/kubeconfig
DEMO_CATALOG ?=

GO_PACKAGES = ./backend/... ./shared/... ./pnetlabaddon/... ./clabgate/...

clean:
	rm -rf ./build

security: clean
	gocritic check -enableAll ./...
	gosec ./...
	golangci-lint run ${GO_PACKAGES} --timeout=10m

pre_commit: clean
	pre-commit run --all-files

generate:
	make -C backend generate
	make -C clabgate generate
	make -C pnetlabaddon generate
	make -C shared generate
	yarn --cwd nextui-dashboard generate

test: clean
	go test -v -coverprofile=coverage.out -covermode=count $(GO_PACKAGES) > tests.out 2>&1; \
	TEST_EXIT_CODE=$$?; \
	cat tests.out; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo "testing failed" >&2; \
		exit $$TEST_EXIT_CODE; \
	fi

generate_certs:
	@sudo openssl req -x509 -nodes -days 3650 -newkey rsa:4096 -keyout ./certs/private.key -out ./certs/public.crt
	@sudo chmod -R 755 ./certs


dev:
	./CI-CD/dev.sh

dev-deps-up:
	$(DEV_COMPOSE) up -d db vault

dev-cluster-up:
	./k8s/local-kind/up.sh --dev

dev-up: dev-deps-up dev-cluster-up
	@echo "Dependencies and kind are ready. Run 'make dev-backend' and 'make dev-clabgate' in separate terminals."

dev-backend:
	cd backend && ./migrate.sh up && go run .

dev-seed:
	cd backend && go run . --demo $(DEMO_CATALOG)

dev-clabgate:
	cd clabgate && \
		SERVER_HOST="$${SERVER_HOST:-0.0.0.0}" \
		KUBECONFIG="$${KUBECONFIG:-$(DEV_KUBECONFIG)}" \
		POD_NAMESPACE="$${POD_NAMESPACE:-cms-labs-system}" \
		go run .

dev-front:
	cd nextui-dashboard && npm run dev

dev-cluster-down:
	./k8s/local-kind/down.sh

dev-down:
	$(DEV_COMPOSE) down
	./k8s/local-kind/down.sh
