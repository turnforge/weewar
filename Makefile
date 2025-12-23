
# Use shell function for runtime evaluation
GO_ROOT=$(shell go env GOROOT)
TINYGO_ROOT=$(shell tinygo env TINYGOROOT 2>/dev/null || echo "")
# Try both common locations for wasm_exec.js (newer Go versions use lib/wasm, older use misc/wasm)
WASM_EXEC_PATH=$(shell find $(GO_ROOT)/lib/wasm $(GO_ROOT)/misc/wasm -name "wasm_exec.js" 2>/dev/null | head -1)

ui:
	cd web ; make build

binlocal: 
	go build -ldflags "$(LDFLAGS)" -o ./bin/weewar ./cmd/backend/*.go

serve:
	WEEWAR_ENV=dev go run cmd/backend/*.go

servelocal:
	WEEWAR_ENV=dev go run cmd/backend/*.go -games_service_be=local -worlds_service_be=local # -gatewayAddress=:6060 -grpcAddress=:7070

vars:
	@echo "GO_ROOT=$(GO_ROOT)"
	@echo "TINYGO_ROOT=$(TINYGO_ROOT)"
	@echo "WASM_EXEC_PATH=$(WASM_EXEC_PATH)"

test:
	@echo "Running tests..."
	go test -cover -coverprofile=coverage.out -coverpkg=./lib/...,./services/... ./tests/... ./cmd/cli/...
	# cd lib && go test -v -cover ./...
	# cd cmd/weewar-cli && go test ./...
	@echo ""
	@echo "Coverage report:"
	go tool cover -func=coverage.out
	@echo ""
	@echo "HTML coverage report: /tmp/coverage.html"
	go tool cover -html=coverage.out -o /tmp/coverage.html
	@echo ""
	@echo "✓ All tests passed"

cli:
	mkdir -p bin
	go build  -o ${GOBIN}/ww cmd/cli/*.go

wasm: # test
	echo "Building WeeWar WASM modules..."
	mkdir -p web/static/wasm
	echo "Building weewar-cli WASM..."
	# GOOS=js GOARCH=wasm go build -o web/static/wasm/weewar-cli.wasm cmd/weewar-wasm/*.go
	echo "Building standard WASM Binary..."
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o web/static/wasm/weewar-cli.wasm ./cmd/weewar-wasm
	echo "Compressing standard WASM Binary..."
	gzip -9 -k -f web/static/wasm/weewar-cli.wasm

tinywasm:
	echo "Building TinyGO WASM Binary..."
	tinygo build -target wasm -o web/static/wasm/weewar-cli-tinygo.wasm ./cmd/weewar-wasm
	echo "Compressing standard WASM Binary..."
	gzip -9 -k -f web/static/wasm/weewar-cli-tinygo.wasm
	echo "Building TinyGO NoDebug WASM Binary..."
	tinygo build -target wasm -no-debug -o web/static/wasm/weewar-cli-tinygo-nodebug.wasm ./cmd/weewar-wasm
	echo "Compressing TinyGo NoDebug WASM Binary..."
	gzip -9 -k -f web/static/wasm/weewar-cli-tinygo-nodebug.wasm

wasmexecjs:
	@echo "Copying wasm_exec.js files..."
	@if [ -n "$(TINYGO_ROOT)" ]; then \
		echo "Copying TinyGo wasm_exec.js from $(TINYGO_ROOT)"; \
		cp "$(TINYGO_ROOT)/targets/wasm_exec.js" web/static/wasm/wasm_exec_tiny.js; \
	else \
		echo "Warning: TinyGo not found, skipping wasm_exec_tiny.js"; \
	fi
	@if [ -n "$(WASM_EXEC_PATH)" ]; then \
		echo "Copying standard Go wasm_exec.js from $(WASM_EXEC_PATH)"; \
		cp "$(WASM_EXEC_PATH)" web/static/wasm/wasm_exec.js; \
	else \
		echo "Warning: wasm_exec.js not found in Go installation at $(GO_ROOT)"; \
	fi
	@echo "File sizes:"
	@du -h web/static/wasm/*.wasm web/static/wasm/*.js 2>/dev/null || true

install-tools:
	@echo "Installing required Go tools..."
	go install golang.org/x/tools/cmd/goyacc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go get  google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get honnef.co/go/tools/cmd/staticcheck
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	go get google.golang.org/genproto/googleapis/api/annotations@latest
	go get github.com/planetscale/vtprotobuf/protohelpers@latest
	go get github.com/planetscale/vtprotobuf@latest
	npm install --force -g  @bufbuild/buf @bufbuild/protobuf @bufbuild/protoc-gen-es @bufbuild/protoc-gen-connect-es
	@echo "✓ Go tools installed"

clean:
	rm -Rf gen
	rm -Rf web/gen

cleanall: clean
	cd protos ; make cleanall

build: down copylinks dockerbuild resymlink
dockerbuild:
	BUILDKIT_PROGRESS=plain docker compose --env-file .env.dev build --no-cache

copylinks:
	rm -Rf locallinks/*
	cp -r ../goutils locallinks/
	cp -r ../goapplib locallinks/
	cp -r ../gocurrent locallinks/
	cp -r ../templar locallinks/
	cp -r ../protoc-gen-dal locallinks/
	cp -r ../oneauth locallinks/
	cp -r ../turnengine locallinks/

resymlink:
	mkdir -p locallinks
	rm -Rf locallinks/*
	cd locallinks && ln -s ../../templar
	cd locallinks && ln -s ../../protoc-gen-dal
	cd locallinks && ln -s ../../goutils
	cd locallinks && ln -s ../../goapplib
	cd locallinks && ln -s ../../gocurrent
	cd locallinks && ln -s ../../oneauth
	cd locallinks && ln -s ../../turnengine

####  Docker related commands

up: ensurenetworks
	docker compose --env-file .env.dev -f docker-compose.yml down
	BUILDKIT_PROGRESS=plain docker compose --env-file .env.dev -f docker-compose.yml up

logs:
	docker compose --env-file .env.dev -f docker-compose.yml logs -f

# Bring everything down
down:
	docker compose --env-file .env.dev -f docker-compose.yml down --remove-orphans
	docker compose --env-file .env.dev -f db-docker-compose.yml down --remove-orphans

# Bring up DB - only brings down DB containers from before - only when we sepearte DB out of docker compose
updb: dbdirs ensurenetworks
	BUILDKIT_PROGRESS=plain docker compose --env-file .env.dev -f db-docker-compose.yml down --remove-orphans
	BUILDKIT_PROGRESS=plain docker compose --env-file .env.dev -f db-docker-compose.yml up -d

dblogs:
	docker compose --env-file .env.dev -f db-docker-compose.yml logs -f

ensurenetworks:
	-docker network create weewarnetwork

dbdirs:
	mkdir -p ./data/pgdata

snap:
	cp -r ~/dev-app-data/weewar/storage/games/testgame snapshotairport

restore:
	cp -r snapshotairport/*.json ~/dev-app-data/weewar/storage/games/testgame/
