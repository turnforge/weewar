# WeeWar - Turn-Based Strategy Game

A modern turn-based strategy game built with Go, featuring hexagonal grid combat, web-based map editor, and CLI gameplay.

## Current Status

🚧 **Active Development** - Major coordinate system refactoring in progress

We are currently migrating from a legacy row/col coordinate system to proper cube coordinates for hexagonal grids. This provides better mathematical foundations and eliminates coordinate conversion bugs.

**Progress**: ~80% complete - Core systems migrated, CLI translation layer in progress

See [COORD_MIGRATION.md](COORD_MIGRATION.md) for detailed progress and technical details.

## Architecture Overview

WeeWar is built with a clean separation of concerns:

- **Core Game Logic** (`lib/`) - Pure Go game engine with hexagonal grid math
- **Web Editor** (`web/`) - Browser-based map editor with real-time rendering
- **CLI Interface** (`cmd/weewar-cli/`) - Command-line game client with chess notation
- **WASM Integration** (`cmd/editor-wasm/`) - WebAssembly bindings for browser

## Key Features

- **Hexagonal Grid Combat** - Proper hex math with cube coordinates
- **Interactive Map Editor** - Visual terrain editing with live preview
- **Chess Notation CLI** - User-friendly A1, B2 position references
- **Multi-format Rendering** - PNG export, canvas rendering, layered composition
- **Asset Management** - Embedded and fetch-based sprite loading

## Key technologies and Stack components:

1. Protos/GRPC services
2. API Fronted by a gateway service
3. Powered by OneAuth for oauth
4. Basic frontend based on Templar go-templates - this can be customized in the future.
5. Tailwind for styling and Typescript for front end.  Add react/vue etc if youd like
6. Webpack for complex pages
7. Many sample page templates, eg ListPages, BorderLayout pages, pagse with DockView etc you can easily copy for other
   parts of your page.

Other backend choices (like datastores) are upto the app/service dev, eg:

1. Which services are to be added.
2. Which backends are to be used (for storage, etc)
3. How to deploy them to specific hosting providers (eg appengine, heroku etc)
4. Selecting frontend frameworks.

## Requirements

1. Basic/Standard Go Tooling:

* Go
* Air (for fast reloads)
* Protobuf
* GRPC
* Buf (for generating artificates from grpc protos)
* Webpack for any complex pages

## Getting Started

1. Clone this Repo

Replace the following variables:

TODO:
1. Common docker compose manifests for packaging for development.
2. Optional k8s configs if needed in the future for testing against cluster deployments

## Conventions:

1. Protos are defined in ./protos folder.  Grpc is our source of truth for everything.  Every other client is generated
   with this.   Connect clients, gateway bindings, TS clients even MCP tool names!
2. Web server, handlers, templates are all defined in the web folder
3. services folder takes care of the core proto service implementation and all the backend heavy lifting
4. The main.go just loads an "App" type that can run a bunch of servers (GrpcServer, WebServer and any others).  Each
   "cluster" has its own server - services has the GrpcServer, web has the WebServer (with grpc bindings).
