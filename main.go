package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

type FileType int

const (
	FileTypeExcel FileType = iota
	FileTypeCSV
	FileTypeUnknown
)

func detectFileType(filename string) FileType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx", ".xls", ".xlsm", ".xltx", ".xltm":
		return FileTypeExcel
	case ".csv":
		return FileTypeCSV
	default:
		return FileTypeUnknown
	}
}

// parseCSV parses a CSV file and returns its contents
func parseCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	return records, nil
}

// writeCSV writes data to a CSV file
func writeCSV(filename string, records [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write all records
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV data: %v", err)
	}

	return nil
}

// extractRegions extracts unique regions (RR00000000) from PSGC data
func extractRegions(records [][]string) [][]string {
	regionMap := make(map[string][]string)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue // Skip header or incomplete rows
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		if psgcCode[2:] == "00000000" {
			regcode := psgcCode[:2]
			regionMap[regcode] = []string{regcode, strings.TrimSpace(row[1])}
		}
	}

	// Convert map to slice
	var result [][]string
	result = append(result, []string{"regcode", "regname"}) // Add header

	for _, region := range regionMap {
		result = append(result, region)
	}

	// Sort by regcode ascending
	sortByCode(result, 0)

	return result
}

// extractProvinces extracts unique provinces from PSGC data
// Returns regcode, provcode (5 digits), and provname
// A place without a province (highly urbanized city, Pateros) is listed as its own province
func extractProvinces(records [][]string) [][]string {
	validProvinces := make(map[string]bool)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		ppp := psgcCode[2:5]
		cc := psgcCode[5:7]
		bbb := psgcCode[7:10]

		if ppp != "000" && cc == "00" && bbb == "000" {
			validProvinces[psgcCode[:5]] = true
		}
	}

	provinceMap := make(map[string][]string)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		name := strings.TrimSpace(row[1])
		ppp := psgcCode[2:5]
		cc := psgcCode[5:7]
		bbb := psgcCode[7:10]

		if ppp != "000" && cc != "00" && bbb == "000" {
			regcode := psgcCode[:2]
			provinceCode := psgcCode[:5]

			if !validProvinces[provinceCode] {
				provinceMap[provinceCode] = []string{regcode, provinceCode, name}
			}
		}
	}

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		name := strings.TrimSpace(row[1])
		ppp := psgcCode[2:5]
		cc := psgcCode[5:7]
		bbb := psgcCode[7:10]

		if ppp != "000" && cc == "00" && bbb == "000" {
			regcode := psgcCode[:2]
			provinceCode := psgcCode[:5]
			provinceMap[provinceCode] = []string{regcode, provinceCode, name}
		}
	}

	var result [][]string
	result = append(result, []string{"regcode", "provcode", "provname"})

	for _, province := range provinceMap {
		result = append(result, province)
	}

	sortByCode(result, 1)

	return result
}

// extractCities extracts unique cities/municipalities from PSGC data
// Returns regcode, provcode (5 digits), citycode (7 digits), and cityname
// A highly urbanized city (RRPPP00000) is listed when barangays sit directly under it
func extractCities(records [][]string) [][]string {
	hasBarangays := make(map[string]bool)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		if psgcCode[7:10] != "000" {
			hasBarangays[psgcCode[:7]] = true
		}
	}

	cityMap := make(map[string][]string)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		name := strings.TrimSpace(row[1])
		ppp := psgcCode[2:5]
		cc := psgcCode[5:7]
		bbb := psgcCode[7:10]

		citycode := psgcCode[:7]

		if ppp != "000" && bbb == "000" && (cc != "00" || hasBarangays[citycode]) {
			regcode := psgcCode[:2]
			provinceCode := psgcCode[:5]

			cityMap[citycode] = []string{regcode, provinceCode, citycode, name}
		}
	}

	var result [][]string
	result = append(result, []string{"regcode", "provcode", "citycode", "cityname"})

	for _, city := range cityMap {
		result = append(result, city)
	}

	sortByCode(result, 2)

	return result
}

// extractBarangays extracts unique barangays from PSGC data
// Returns regcode, provcode (5 digits), citycode (7 digits), barangaycode (10 digits), and barangayname
func extractBarangays(records [][]string) [][]string {
	barangayMap := make(map[string][]string)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		name := strings.TrimSpace(row[1])
		bbb := psgcCode[7:10]

		if bbb != "000" {
			regcode := psgcCode[:2]
			provinceCode := psgcCode[:5]
			citycode := psgcCode[:7]
			barangaycode := psgcCode

			barangayMap[barangaycode] = []string{regcode, provinceCode, citycode, barangaycode, name}
		}
	}

	var result [][]string
	result = append(result, []string{"regcode", "provcode", "citycode", "barangaycode", "barangayname"})

	for _, barangay := range barangayMap {
		result = append(result, barangay)
	}

	// Sort by barangaycode ascending
	sortByCode(result, 3)

	return result
}

// cleanPSGC converts raw hierarchical PSGC data to cleaned flattened format
// Uses pure 10-digit PSGC code structure parsing:
//   - RR00000000: Region
//   - RRPPP0000 (PPP≠000): Province/HUC
//   - RRPPPCC000 (CC≠00): City/Municipality
//   - RRPPPCCBBB (BBB≠000): Barangay
func cleanPSGC(records [][]string) [][]string {
	validProvinces := make(map[string]bool)

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		if psgcCode[2:5] != "000" && psgcCode[5:7] == "00" && psgcCode[7:10] == "000" {
			validProvinces[psgcCode[:5]] = true
		}
	}

	var regionCode, provinceCode, cityCode string
	var regionName, provinceName, cityName string

	var result [][]string
	result = append(result, []string{
		"Region Code", "Region Name", "Province Code", "Province Name",
		"City/Municipality Code", "City/Municipality Name", "Barangay Code", "Barangay Name",
	})

	for i, row := range records {
		if i == 0 || len(row) < 2 {
			continue
		}

		psgcCode := normalizePSGC(row[0])
		if len(psgcCode) != 10 {
			continue
		}

		name := strings.TrimSpace(row[1])
		ppp := psgcCode[2:5]
		cc := psgcCode[5:7]
		bbb := psgcCode[7:10]

		if ppp == "000" && cc == "00" && bbb == "000" {
			regionCode = psgcCode
			regionName = name
			provinceCode = ""
			provinceName = ""
			cityCode = ""
			cityName = ""
		} else if ppp != "000" && cc == "00" && bbb == "000" {
			provinceCode = psgcCode
			provinceName = name
			cityCode = ""
			cityName = ""
		} else if ppp != "000" && cc != "00" && bbb == "000" {
			cityCode = psgcCode
			cityName = name

			parentProv := psgcCode[:5]
			if !validProvinces[parentProv] {
				provinceCode = parentProv + "00000"
				provinceName = name
			}
		} else if ppp != "000" && bbb != "000" {
			// Barangays of a highly urbanized city sit directly under its province-level code
			if cc == "00" {
				cityCode = provinceCode
				cityName = provinceName
			}

			result = append(result, []string{
				regionCode,
				regionName,
				provinceCode,
				provinceName,
				cityCode,
				cityName,
				psgcCode,
				name,
			})
		}
	}

	return result
}

func normalizePSGC(rawCode string) string {
	code := strings.TrimSpace(rawCode)
	if code == "" {
		return "" // Padding a blank code would turn it into a region
	}
	if strings.Contains(code, ".") {
		code = strings.Split(code, ".")[0]
	}
	for len(code) < 10 {
		code = "0" + code
	}
	return code
}

// readExcelSheet reads data from a specific sheet in an Excel file
func readExcelSheet(filename, sheetName string) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close Excel file: %v", err)
		}
	}()

	// Get all rows from the specified sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %v", sheetName, err)
	}

	return rows, nil
}

// sortByCode sorts the records by the given code column in ascending order
// Skips the header row (index 0)
func sortByCode(records [][]string, column int) {
	if len(records) <= 1 {
		return // No data to sort or only header
	}

	sort.Slice(records[1:], func(i, j int) bool {
		return records[i+1][column] < records[j+1][column]
	})
}

// printSample displays a sample of the parsed data
func printSample(records [][]string) {
	fmt.Printf("Total rows: %d\n", len(records))

	// Print first few rows as sample
	maxRows := 5
	if len(records) < maxRows {
		maxRows = len(records)
	}

	fmt.Println("\nSample data:")
	for i := 0; i < maxRows; i++ {
		fmt.Printf("Row %d: %v\n", i+1, records[i])
	}

	if len(records) > maxRows {
		fmt.Printf("... and %d more rows\n", len(records)-maxRows)
	}
}

func main() {
	// Define command line flags
	outputFile := flag.String("output", "", "Output CSV file (optional)")
	extractRegionsFlag := flag.Bool("regions", false, "Extract unique regions (regcode, regname) from PSGC data")
	extractProvincesFlag := flag.Bool("provinces", false, "Extract unique provinces (regcode, provcode, provname) from PSGC data")
	extractCitiesFlag := flag.Bool("cities", false, "Extract unique cities/municipalities (regcode, provcode, citycode, cityname) from PSGC data")
	extractBarangaysFlag := flag.Bool("barangays", false, "Extract unique barangays (regcode, provcode, citycode, barangaycode, barangayname) from PSGC data")
	extractAllFlag := flag.Bool("all", false, "Extract all geographic levels (regions, provinces, cities, barangays) from PSGC data")
	cleanPSGCFlag := flag.Bool("clean", false, "Convert raw hierarchical PSGC data to cleaned flattened format")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run main.go [options] <input_file>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("Supported input formats: Excel (.xlsx, .xls, .xlsm, .xltx, .xltm) and CSV (.csv)")
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	// Options may also follow the input file
	flag.CommandLine.Parse(flag.Args()[1:])
	if flag.NArg() > 0 {
		log.Fatalf("Unexpected argument: %s", flag.Arg(0))
	}

	modes := 0
	for _, set := range []bool{*extractRegionsFlag, *extractProvincesFlag, *extractCitiesFlag, *extractBarangaysFlag, *extractAllFlag, *cleanPSGCFlag} {
		if set {
			modes++
		}
	}
	if modes > 1 {
		log.Fatal("Use only one of -regions, -provinces, -cities, -barangays, -all, or -clean")
	}

	// Detect file type based on extension
	fileType := detectFileType(inputFile)

	var records [][]string
	var err error

	// Check if any extraction flags are set
	extractionRequested := *extractRegionsFlag || *extractProvincesFlag || *extractCitiesFlag || *extractBarangaysFlag || *extractAllFlag

	// Parse the input file
	switch fileType {
	case FileTypeExcel:
		// For Excel files, if extraction is requested, read directly from "PSGC" sheet
		// Otherwise, parse all sheets (for backward compatibility with existing workflows)
		if extractionRequested && !*cleanPSGCFlag {
			records, err = readExcelSheet(inputFile, "PSGC")
			if err != nil {
				log.Fatalf("Failed to read PSGC sheet from Excel file: %v", err)
			}
			fmt.Printf("Read PSGC sheet from Excel file: %s\n", inputFile)
		} else {
			records, err = parseExcel(inputFile)
			if err != nil {
				log.Fatalf("Failed to parse Excel file: %v", err)
			}
			fmt.Printf("Parsed Excel file: %s\n", inputFile)
		}
	case FileTypeCSV:
		records, err = parseCSV(inputFile)
		if err != nil {
			log.Fatalf("Failed to parse CSV file: %v", err)
		}
		fmt.Printf("Parsed CSV file: %s\n", inputFile)
	default:
		log.Fatalf("Unsupported file type. Supported formats: Excel (.xlsx, .xls, .xlsm, .xltx, .xltm) and CSV (.csv)")
	}

	// A cleaned file holds each barangay with its parents, so split it back into code and name rows
	if extractionRequested && len(records) > 0 && len(records[0]) == 8 && records[0][0] == "Region Code" {
		split := [][]string{{"10-digit PSGC", "Name"}}
		for _, row := range records[1:] {
			for column := 0; column < len(row); column += 2 {
				split = append(split, []string{row[column], row[column+1]})
			}
		}
		records = split
	}

	// Extract regions if requested
	if *extractRegionsFlag {
		records = extractRegions(records)
		fmt.Println("Extracted unique regions from data")
	}

	// Extract provinces if requested
	if *extractProvincesFlag {
		records = extractProvinces(records)
		fmt.Println("Extracted unique provinces from data")
	}

	// Extract cities if requested
	if *extractCitiesFlag {
		records = extractCities(records)
		fmt.Println("Extracted unique cities/municipalities from data")
	}

	// Extract barangays if requested
	if *extractBarangaysFlag {
		records = extractBarangays(records)
		fmt.Println("Extracted unique barangays from data")
	}

	// Clean PSGC data if requested
	if *cleanPSGCFlag {
		if fileType == FileTypeExcel {
			// For Excel files, read specifically from "PSGC" sheet
			excelErr := error(nil)
			records, excelErr = readExcelSheet(inputFile, "PSGC")
			if excelErr != nil {
				log.Fatalf("Failed to read PSGC sheet from Excel file: %v", excelErr)
			}
			fmt.Printf("Read PSGC sheet from Excel file: %s\n", inputFile)
		}
		records = cleanPSGC(records)
		fmt.Println("Cleaned PSGC data to flattened format")
	}

	// Extract all geographic levels if requested
	if *extractAllFlag {
		fmt.Println("Extracting all geographic levels...")

		// Create output directory
		if err := os.MkdirAll("output", 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		// Extract regions
		regions := extractRegions(records)
		if err := writeCSV("output/rregion.csv", regions); err != nil {
			log.Fatalf("Failed to write regions CSV: %v", err)
		}
		fmt.Printf("Created output/rregion.csv (%d regions)\n", len(regions)-1)

		// Extract provinces
		provinces := extractProvinces(records)
		if err := writeCSV("output/rprovince.csv", provinces); err != nil {
			log.Fatalf("Failed to write provinces CSV: %v", err)
		}
		fmt.Printf("Created output/rprovince.csv (%d provinces)\n", len(provinces)-1)

		// Extract cities
		cities := extractCities(records)
		if err := writeCSV("output/rcity.csv", cities); err != nil {
			log.Fatalf("Failed to write cities CSV: %v", err)
		}
		fmt.Printf("Created output/rcity.csv (%d cities/municipalities)\n", len(cities)-1)

		// Extract barangays
		barangays := extractBarangays(records)
		if err := writeCSV("output/rbarangay.csv", barangays); err != nil {
			log.Fatalf("Failed to write barangays CSV: %v", err)
		}
		fmt.Printf("Created output/rbarangay.csv (%d barangays)\n", len(barangays)-1)

		fmt.Println("All geographic levels extracted successfully to 'output' directory!")
		return
	}

	// Create output directory for individual extractions
	if *extractRegionsFlag || *extractProvincesFlag || *extractCitiesFlag || *extractBarangaysFlag || *cleanPSGCFlag {
		if err := os.MkdirAll("output", 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	// Set default output file for region extraction
	if *extractRegionsFlag && *outputFile == "" {
		*outputFile = "output/rregion.csv"
	}

	// Set default output file for province extraction
	if *extractProvincesFlag && *outputFile == "" {
		*outputFile = "output/rprovince.csv"
	}

	// Set default output file for city extraction
	if *extractCitiesFlag && *outputFile == "" {
		*outputFile = "output/rcity.csv"
	}

	// Set default output file for barangay extraction
	if *extractBarangaysFlag && *outputFile == "" {
		*outputFile = "output/rbarangay.csv"
	}

	// Set default output file for PSGC cleaning
	if *cleanPSGCFlag && *outputFile == "" {
		*outputFile = "output/psgc_cleaned_go.csv"
	}

	// Output to CSV file if specified
	if *outputFile != "" {
		if err := writeCSV(*outputFile, records); err != nil {
			log.Fatalf("Failed to write CSV output: %v", err)
		}
		fmt.Printf("Data written to CSV file: %s\n", *outputFile)
	}

	// Display sample data to console
	printSample(records)
}

// parseExcel parses an Excel file and returns its contents as CSV-compatible data
func parseExcel(filename string) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close Excel file: %v", err)
		}
	}()

	// Get all sheet names
	sheetNames := f.GetSheetList()
	var allRecords [][]string

	// Process each sheet
	for _, sheetName := range sheetNames {
		// Get all rows in the sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Printf("Failed to get rows from sheet %s: %v", sheetName, err)
			continue
		}

		// Add sheet name as first column for each row
		for _, row := range rows {
			record := append([]string{sheetName}, row...)
			allRecords = append(allRecords, record)
		}
	}

	return allRecords, nil
}
