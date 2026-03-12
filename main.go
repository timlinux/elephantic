package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twpayne/go-geos"
)

//go:embed countries.gpkg
var countriesData []byte

type Country struct {
	ISOCode  string
	Geometry *geos.Geom
	Prepared *geos.PrepGeom
}

type PointCountResult struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
	Errors int            `json:"errors"`
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type     string          `json:"type"`
	Geometry GeoJSONGeometry `json:"geometry"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

var (
	countries     []Country
	countriesOnce sync.Once
	gpkgTempPath  string
)

func loadCountries() error {
	var loadErr error
	countriesOnce.Do(func() {
		// Write embedded data to temp file (sqlite needs file path)
		tmpFile, err := os.CreateTemp("", "countries-*.gpkg")
		if err != nil {
			loadErr = fmt.Errorf("failed to create temp file: %w", err)
			return
		}
		gpkgTempPath = tmpFile.Name()

		if _, err := tmpFile.Write(countriesData); err != nil {
			loadErr = fmt.Errorf("failed to write temp file: %w", err)
			return
		}
		tmpFile.Close()

		// Open geopackage
		db, err := sql.Open("sqlite3", gpkgTempPath)
		if err != nil {
			loadErr = fmt.Errorf("failed to open geopackage: %w", err)
			return
		}
		defer db.Close()

		// Query countries - using iso_a3 for 3-letter codes
		rows, err := db.Query(`
			SELECT iso_a3, geom
			FROM countries
			WHERE iso_a3 IS NOT NULL AND iso_a3 != '' AND iso_a3 != '-99'
		`)
		if err != nil {
			loadErr = fmt.Errorf("failed to query countries: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var isoCode string
			var geomBlob []byte

			if err := rows.Scan(&isoCode, &geomBlob); err != nil {
				continue
			}

			// Parse GeoPackage geometry (skip header for standard WKB)
			if len(geomBlob) < 8 {
				continue
			}

			// GeoPackage geometry header: magic (2) + version (1) + flags (1) + srs_id (4) + envelope (variable)
			// Find where WKB starts by checking the flags byte
			flags := geomBlob[3]
			envelopeType := (flags >> 1) & 0x07
			headerSize := 8 // minimum header
			switch envelopeType {
			case 1:
				headerSize += 32 // 4 doubles (minx, maxx, miny, maxy)
			case 2:
				headerSize += 48 // 6 doubles (+ minz, maxz)
			case 3:
				headerSize += 48 // 6 doubles (+ minm, maxm)
			case 4:
				headerSize += 64 // 8 doubles (+ minz, maxz, minm, maxm)
			}

			if len(geomBlob) <= headerSize {
				continue
			}

			wkb := geomBlob[headerSize:]
			geom, err := geos.NewGeomFromWKB(wkb)
			if err != nil {
				continue
			}

			prepared := geom.Prepare()

			countries = append(countries, Country{
				ISOCode:  isoCode,
				Geometry: geom,
				Prepared: prepared,
			})
		}

		if len(countries) == 0 {
			loadErr = fmt.Errorf("no countries loaded")
		}
	})
	return loadErr
}

func countPointsPerCountry(geojsonInput string) (*PointCountResult, error) {
	if err := loadCountries(); err != nil {
		return nil, fmt.Errorf("failed to load countries: %w", err)
	}

	// Parse GeoJSON
	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal([]byte(geojsonInput), &fc); err != nil {
		return nil, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}

	// Extract points
	var points [][]float64
	for _, feature := range fc.Features {
		if feature.Geometry.Type == "Point" && len(feature.Geometry.Coordinates) >= 2 {
			points = append(points, feature.Geometry.Coordinates)
		}
	}

	if len(points) == 0 {
		return &PointCountResult{
			Counts: make(map[string]int),
			Total:  0,
			Errors: 0,
		}, nil
	}

	// Parallel processing with worker pool
	numWorkers := runtime.NumCPU()
	pointChan := make(chan []float64, len(points))
	resultChan := make(chan string, len(points))
	var errorCount int64

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for coords := range pointChan {
				point := geos.NewPointFromXY(coords[0], coords[1])
				if point == nil {
					atomic.AddInt64(&errorCount, 1)
					resultChan <- ""
					continue
				}

				found := ""
				for _, country := range countries {
					// Use prepared geometry for faster contains check
					if country.Prepared.ContainsProperly(point) {
						found = country.ISOCode
						break
					}
				}
				point.Destroy()
				resultChan <- found
			}
		}()
	}

	// Send points to workers
	go func() {
		for _, p := range points {
			pointChan <- p
		}
		close(pointChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	counts := make(map[string]int)
	for isoCode := range resultChan {
		if isoCode != "" {
			counts[isoCode]++
		}
	}

	return &PointCountResult{
		Counts: counts,
		Total:  len(points),
		Errors: int(errorCount),
	}, nil
}

func toolResultError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}},
		IsError: true,
	}
}

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"elephantic",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register the count_points_by_country tool
	tool := mcp.NewTool("count_points_by_country",
		mcp.WithDescription("Count GeoJSON points by country using Natural Earth boundaries. Returns ISO country codes with point counts."),
		mcp.WithString("geojson",
			mcp.Required(),
			mcp.Description("GeoJSON FeatureCollection containing Point geometries"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		geojson, ok := request.Params.Arguments["geojson"].(string)
		if !ok {
			return toolResultError("geojson parameter must be a string"), nil
		}

		result, err := countPointsPerCountry(geojson)
		if err != nil {
			return toolResultError(err.Error()), nil
		}

		jsonResult, err := json.Marshal(result)
		if err != nil {
			return toolResultError("failed to serialize result"), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Start stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
