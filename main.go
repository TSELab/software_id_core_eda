package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TSELab/software_id_core_eda/orationis"
	"github.com/guacsec/sw-id-core/coordinates"
)

func main() {
	sbomDir := "/Users/ravle/guac_analytics/sboms-cyclonedx"

	validatedFile, err := os.Create("validated_coordinates.csv")
	if err != nil {
		log.Fatalf("Could not create validated_coordinates.csv: %v", err)
	}
	defer validatedFile.Close()
	validWriter := csv.NewWriter(validatedFile)
	defer validWriter.Flush()

	failuresFile, err := os.Create("failures.csv")
	if err != nil {
		log.Fatalf("Could not create failures.csv: %v", err)
	}
	defer failuresFile.Close()
	failWriter := csv.NewWriter(failuresFile)
	defer failWriter.Flush()

	validWriter.Write([]string{
		"name", "purl", "source",
		"valid_coordinate", "coordinate_type", "coordinate_revision",
	})
	failWriter.Write([]string{
		"file", "purl", "error",
	})

	var totalSBOMs, parsedSBOMs, failedSBOMs, totalIdentifiers, totalValid int

	packageTypeCounts := make(map[string]int)
	basePurlCounts := make(map[string]int)
	fullPurlCounts := make(map[string]int)
	revisionCounts := make(map[string]int)
	sourceCounts := make(map[string]int)
	sizeBuckets := make(map[string]int)

	err = filepath.Walk(sbomDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		totalSBOMs++
		fileName := filepath.Base(path)
		fileValid := 0

		identifiers, err := orationis.ParseCycloneDX(path)
		if err != nil {
			fmt.Printf("[ERROR] Failed to parse %s: %v\n", path, err)
			failWriter.Write([]string{fileName, "", err.Error()})
			failedSBOMs++
			return nil
		}

		// Reject SBOMs with missing metadata.component.name
		if len(identifiers) == 0 || identifiers[0].Name == "MISSING" {
			failWriter.Write([]string{fileName, "", "missing metadata.component.name"})
			failedSBOMs++
			return nil
		}

		parsedSBOMs++
		totalIdentifiers += len(identifiers)

		for _, id := range identifiers {
			fullPurl := id.Purl

			coord, err := coordinates.ConvertPurlToCoordinate(fullPurl)
			if err != nil {
				failWriter.Write([]string{fileName, fullPurl, err.Error()})
				continue
			}

			validWriter.Write([]string{
				id.Name,
				fullPurl,
				id.Source,
				"true",
				coord.CoordinateType,
				coord.Revision,
			})

			basePurl := stripQualifiers(fullPurl)
			basePurlCounts[basePurl]++
			fullPurlCounts[fullPurl]++

			packageType := coord.CoordinateType
			if packageType == "" {
				packageType = "unknown"
			}
			packageTypeCounts[packageType]++
			revisionCounts[coord.Revision]++
			sourceCounts[id.Source]++

			fileValid++
			totalValid++
		}

		switch {
		case fileValid <= 10:
			sizeBuckets["1–10"]++
		case fileValid <= 50:
			sizeBuckets["11–50"]++
		case fileValid <= 100:
			sizeBuckets["51–100"]++
		default:
			sizeBuckets["100+"]++
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking SBOM directory: %v", err)
	}

	validWriter.Flush()
	failWriter.Flush()

	summaryFile, err := os.Create("summary_stats.txt")
	if err != nil {
		log.Fatalf("Could not create summary_stats.txt: %v", err)
	}
	defer summaryFile.Close()

	writeSummary(summaryFile, totalSBOMs, parsedSBOMs, failedSBOMs, totalIdentifiers, totalValid,
		packageTypeCounts, basePurlCounts, fullPurlCounts, revisionCounts, sourceCounts, sizeBuckets)

	fmt.Println("\n✔ Output files generated:")
	fmt.Println(" - validated_coordinates.csv")
	fmt.Println(" - failures.csv")
	fmt.Println(" - summary_stats.txt")
}

func writeSummary(f *os.File, totalSBOMs, parsedSBOMs, failedSBOMs, totalIdentifiers, totalValid int,
	packageTypeCounts, basePurlCounts, fullPurlCounts, revisionCounts, sourceCounts, sizeBuckets map[string]int) {

	fmt.Fprintf(f, "=== SUMMARY STATS ===\n")
	fmt.Fprintf(f, "Total SBOM files scanned: %d\n", totalSBOMs)
	fmt.Fprintf(f, "Successfully parsed SBOMs: %d\n", parsedSBOMs)
	fmt.Fprintf(f, "Failed SBOMs: %d\n", failedSBOMs)
	fmt.Fprintf(f, "Total identifiers parsed: %d\n", totalIdentifiers)
	fmt.Fprintf(f, "Total valid coordinates: %d\n\n", totalValid)

	fmt.Fprintf(f, "Counts by package type (sorted):\n")
	writeSortedCounts(f, packageTypeCounts)

	fmt.Fprintf(f, "\nTop 10 most common base purls (no qualifiers):\n")
	writeTopN(f, basePurlCounts, 10)

	fmt.Fprintf(f, "\nTop 20 most common unique full purls (from validated_coordinates.csv):\n")
	writeTopN(f, fullPurlCounts, 20)

	fmt.Fprintf(f, "\nTop 10 most common coordinate revisions:\n")
	writeTopN(f, revisionCounts, 10)

	fmt.Fprintf(f, "\nTop 10 most common sources:\n")
	writeTopN(f, sourceCounts, 10)

	fmt.Fprintf(f, "\nSBOM size distribution (by valid identifiers):\n")
	for _, b := range []string{"1–10", "11–50", "51–100", "100+"} {
		fmt.Fprintf(f, "%-7s : %d files\n", b, sizeBuckets[b])
	}
}

func stripQualifiers(purl string) string {
	if idx := strings.IndexAny(purl, "?#"); idx != -1 {
		return purl[:idx]
	}
	return purl
}

func writeSortedCounts(f *os.File, counts map[string]int) {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(f, "%-20s : %d\n", k, counts[k])
	}
}

func writeTopN(f *os.File, counts map[string]int, N int) {
	type pair struct {
		Key   string
		Count int
	}
	var list []pair
	for k, v := range counts {
		list = append(list, pair{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	for i := 0; i < N && i < len(list); i++ {
		shortKey := list[i].Key
		if len(shortKey) > 100 {
			shortKey = shortKey[:100] + "..."
		}
		fmt.Fprintf(f, "%4d %s\n", list[i].Count, shortKey)
	}
}
