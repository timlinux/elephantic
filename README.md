# Elephantic

A high-performance MCP (Model Context Protocol) server that counts GeoJSON points by country using embedded Natural Earth boundaries.

## Features

- Embedded Natural Earth country boundaries (no external data files needed)
- Highly parallelized point-in-polygon processing
- Scales to thousands of simultaneous users
- Static binary with GEOS for fast spatial operations
- Returns ISO country codes with point counts

## Quick Start

### Prerequisites

- Nix with flakes enabled, OR
- Go 1.22+ with GEOS development libraries

### Building

**With Nix (recommended):**
```bash
nix develop
go mod tidy
go build -o elephantic .
```

**Without Nix:**
```bash
# Install GEOS (Ubuntu/Debian)
sudo apt install libgeos-dev

# Install GEOS (macOS)
brew install geos

# Build
CGO_ENABLED=1 go build -o elephantic .
```

### Providing Country Data

Place your Natural Earth countries GeoPackage as `countries.gpkg` in the project root before building. The GeoPackage should have a table named `countries` with an `iso_a3` column for ISO 3166-1 alpha-3 country codes.

Download Natural Earth data:
```bash
wget https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_0_countries.zip
unzip ne_10m_admin_0_countries.zip
ogr2ogr -f GPKG countries.gpkg ne_10m_admin_0_countries.shp -nln countries
```

## MCP Configuration

### Claude Desktop / Claude Code

Add to `~/.config/claude/claude_desktop_config.json` (Linux) or `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "elephantic": {
      "command": "/path/to/elephantic"
    }
  }
}
```

For Claude Code, add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "elephantic": {
      "command": "/path/to/elephantic"
    }
  }
}
```

### ChatGPT (via MCP Bridge)

ChatGPT doesn't natively support MCP, but you can use an MCP-to-OpenAI bridge:

```bash
# Using mcp-bridge (if available)
mcp-bridge --server /path/to/elephantic --port 8080
```

Then configure ChatGPT to use the bridge endpoint as a custom function.

### Gemini CLI

Add to your Gemini CLI configuration:

```json
{
  "tools": {
    "elephantic": {
      "type": "mcp",
      "command": "/path/to/elephantic"
    }
  }
}
```

## Usage

Once configured, you can ask your AI assistant:

> "Count how many of these points are in each country: [GeoJSON data]"

The tool accepts a GeoJSON FeatureCollection with Point geometries and returns:

```json
{
  "counts": {
    "US": 45,
    "CA": 12,
    "MX": 8
  },
  "total": 65,
  "errors": 0
}
```

## API

### Tool: `count_points_by_country`

**Input:**
- `geojson` (string, required): GeoJSON FeatureCollection containing Point geometries

**Output:**
- `counts`: Object mapping ISO country codes to point counts
- `total`: Total number of points processed
- `errors`: Number of points that failed to process

## Performance

- Uses GEOS prepared geometries for O(log n) point-in-polygon checks
- Worker pool scales to available CPU cores
- Thread-safe with per-worker GEOS contexts
- Countries loaded once and cached

## License

MIT License

---

Made with love by [Kartoza](https://kartoza.com) | [Donate](https://github.com/sponsors/kartoza) | [GitHub](https://github.com/timlinux/elephantic)
