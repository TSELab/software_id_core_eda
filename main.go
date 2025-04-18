package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/TSELab/software_id_core_eda/orationis"
	"github.com/guacsec/sw-id-core/coordinates"
)

func main() {
	sbomDir := "/Users/ravle/guac_analytics/sboms-cyclonedx"
	outputFile, err := os.Create("validated_coordinates.csv")
	if err != nil {
		log.Fatalf("Could not create output CSV: %v", err)
	}
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{
		"name", "version", "purl", "type", "bomRef", "source",
		"valid_coordinate", "coordinate_type", "coordinate_revision",
	})

	var totalSBOMs int
	var parsedSBOMs int
	var failedSBOMs int
	var totalIdentifiers int

	// Walk through directory and process each .json SBOM file
	err = filepath.Walk(sbomDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // if the file can’t be accessed, skip it
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		totalSBOMs++

		identifiers, err := orationis.ParseCycloneDX(path)
		if err != nil {
			fmt.Printf("[ERROR] Failed to parse %s: %v\n", path, err)
			failedSBOMs++
			return nil // continue to next file
		}

		parsedSBOMs++
		totalIdentifiers += len(identifiers)

		for _, id := range identifiers {
			valid := "false"
			coordType := ""
			revision := ""

			coord, err := coordinates.ConvertPurlToCoordinate(id.Purl)
			if err == nil {
				valid = "true"
				coordType = coord.CoordinateType
				revision = coord.Revision
			}

			writer.Write([]string{
				id.Name,
				id.Version,
				id.Purl,
				id.Type,
				id.BomRef,
				id.Source,
				valid,
				coordType,
				revision,
			})
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking through SBOM directory: %v", err)
	}

	fmt.Println("CSV written to validated_coordinates.csv")
	fmt.Println("Stats:")
	fmt.Printf("- Total SBOM files scanned: %d\n", totalSBOMs)
	fmt.Printf("- Successfully parsed SBOMs: %d\n", parsedSBOMs)
	fmt.Printf("- Failed SBOMs: %d\n", failedSBOMs)
	fmt.Printf("- Identifiers parsed: %d\n", totalIdentifiers)
}