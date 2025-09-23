
GO_ROOT=$(go env GOROOT)
WASM_EXEC_PATH=$(find `go env GOROOT` -name "wasm_exec.js" 2>/dev/null | head -1)

buildweb:
	cd web ; npm run build

binlocal: 
	go build -ldflags "$(LDFLAGS)" -o ./bin/weewar ./main.go

vars:
	go env GOROOT
	find `go env GOROOT` -name "wasm_exec.js" 2>/dev/null | head -1
	echo GO_ROOT=${GO_ROOT}
	echo WASM_EXEC_PATH=${WASM_EXEC_PATH}

test:
	cd services && go test . -cover
	cd lib && go test . -cover
	# cd cmd/weewar-cli && go test ./...

buf: ensureenv clean
	buf generate
	goimports -w `find gen | grep "\.go"`

cli:
	mkdir -p bin
	go build  -o ./bin/weewar-cli cmd/weewar-cli/*.go
	# go build  -o ./bin/weewar-convert cmd/weewar-convert/*.go

wasm: 
	echo "Building WeeWar WASM modules..."
	mkdir -p web/static/wasm
	echo "Building weewar-cli WASM..."
	GOOS=js GOARCH=wasm go build -o web/static/wasm/weewar-cli.wasm cmd/weewar-wasm/*.go
	echo "Copying wasm_exec.js..."

wasmexecjs:
	GO_ROOT = $(go env GOROOT)
	WASM_EXEC_PATH = $(find "$GO_ROOT" -name "wasm_exec.js" 2>/dev/null | head -1)
	cp "${WASM_EXEC_PATH}" web/static/wasm/
	# echo "Warning: wasm_exec.js not found in Go installation"
	echo "File sizes:"
	du -h web/static/wasm/*.wasm web/static/wasm/*.js

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
	npm install --force -g  @bufbuild/buf @bufbuild/protobuf @bufbuild/protoc-gen-es @bufbuild/protoc-gen-connect-es
	@echo "✓ Go tools installed"

clean:
	rm -Rf gen
	rm -Rf web/gen
	rm -f buf.lock

cleanall: clean remove-proto-symlinks
	rm -f buf.yaml
	rm -f buf.gen.yaml

setupdev: cleanall symlink-protos
	ln -s buf.gen.yaml.dev buf.gen.yaml
	ln -s buf.yaml.dev buf.yaml

setupprod: cleanall remove-proto-symlinks
	ln -s buf.gen.yaml.prod buf.gen.yaml
	ln -s buf.yaml.prod buf.yaml

ensureenv:
	@test -f buf.yaml && test -f buf.gen.yaml && echo "buf.yaml does not exist.  Run 'make setupdev' or 'make setupprod' to setup your environment..."

# Create symlink to wasmjs annotations for development
symlink-protos: remove-proto-symlinks
	echo "Creating turnengine symlink for development..."
	ln -s ../../../engine/protos/turnengine protos/turnengine

# Remove symlink (for switching back to production mode)
remove-proto-symlinks:
	echo "Removing turnengine proto symlink..."
	rm -Rf protos/wasmjs protos/turnengine
