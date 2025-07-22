
GO_ROOT=$(go env GOROOT)
WASM_EXEC_PATH=$(find `go env GOROOT` -name "wasm_exec.js" 2>/dev/null | head -1)

buildweb:
	cd web ; npm run build

binlocal: 
	go build -ldflags "$(LDFLAGS)" -o /tmp/weewar ./main.go

vars:
	go env GOROOT
	find `go env GOROOT` -name "wasm_exec.js" 2>/dev/null | head -1
	echo GO_ROOT=${GO_ROOT}
	echo WASM_EXEC_PATH=${WASM_EXEC_PATH}

test:
	cd lib && go test -v ./...
	cd cmd/weewar-cli && go test -v ./...

buf:
	buf generate

cli:
	cd cmd/weewar-cli && go build .

wasm:
	echo "Building WeeWar WASM modules..."
	mkdir -p web/static/wasm
	echo "Building weewar-cli WASM..."
	GOOS=js GOARCH=wasm go build -o web/static/wasm/weewar-cli.wasm cmd/weewar-wasm/*.go
	# echo "Building map editor WASM..."
	# GOOS=js GOARCH=wasm go build -o web/static/wasm/editor.wasm cmd/editor-wasm/*.go
	echo "Copying wasm_exec.js..."

wasmexecjs:
	GO_ROOT = $(go env GOROOT)
	WASM_EXEC_PATH = $(find "$GO_ROOT" -name "wasm_exec.js" 2>/dev/null | head -1)
	cp "${WASM_EXEC_PATH}" web/static/wasm/
	# echo "Warning: wasm_exec.js not found in Go installation"
	echo "File sizes:"
	du -h web/static/wasm/*.wasm web/static/wasm/*.js
