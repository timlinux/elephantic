# Elephantic Package Architecture

## Dependencies

### Core Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/mark3labs/mcp-go` | v0.17.0 | MCP protocol implementation for Go. Provides server framework and tool registration. |
| `github.com/mattn/go-sqlite3` | v1.14.22 | SQLite driver with CGO. Required for reading embedded GeoPackage files. |
| `github.com/twpayne/go-geos` | v0.18.0 | GEOS bindings for Go. Provides high-performance spatial operations including point-in-polygon. |

### System Dependencies

| Package | Purpose |
|---------|---------|
| `geos` | GEOS C library for geometric operations |
| `sqlite` | SQLite library for GeoPackage reading |
| `pkg-config` | Build tool for locating C libraries |

## Package Structure

```
elephantic/
├── main.go           # Entry point, MCP server setup, tool handler
├── go.mod            # Go module definition
├── go.sum            # Dependency checksums
├── countries.gpkg    # Embedded Natural Earth data
└── flake.nix         # Nix development environment
```

## Build Requirements

- Go 1.22 or later
- CGO enabled (`CGO_ENABLED=1`)
- GEOS development headers
- SQLite development headers

## Embedded Data

The `countries.gpkg` file is embedded at compile time using Go's `//go:embed` directive. This creates a fully self-contained binary with no external data dependencies.
