FROM golang:latest
# FROM panyam/golangdev:latest

WORKDIR /app

COPY locallinks ./locallinks
COPY Makefile go.mod go.sum ./
RUN mkdir -p cmd clients datastore services gen locallinks
COPY cmd ./cmd
COPY assets ./assets
COPY gen ./gen
COPY locallinks ./locallinks
COPY services ./services
COPY utils ./utils

RUN mkdir -p web web/server web/gen web/static
COPY web/server ./web/server ./web/static
COPY web/gen ./web/gen
RUN ls -al

RUN go mod download

WORKDIR /app/cmd/backend

# Uncomment below line to eee if the build works as a test
# Usually this will be useful if you have added or removed folders
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/main main.go

CMD ["go", "run", "main.go"]
