# PSGC Parser

A Go-based parser for Philippine Standard Geographic Code (PSGC) data from the official publication datafile. Supports Excel (.xlsx) and CSV formats for cleaning and extracting hierarchical geographic data.

## Features

- **Data Cleaning**: Convert raw hierarchical PSGC data to flattened CSV format
- **Hierarchical Extraction**: Extract regions, provinces, cities/municipalities, and barangays
- **Multiple Formats**: Supports Excel (.xlsx) and CSV files
- **Batch Processing**: Extract all geographic levels at once

## Requirements

### System Requirements
- **OS**: Linux, macOS, or Windows
- **Go**: Version 1.26 or higher (Go 1.21 and later download it automatically unless `GOTOOLCHAIN` is changed)

### Dependencies
The project requires the following Go package:
- `github.com/xuri/excelize/v2` - For reading Excel files

### Input File Requirements
- **Format**: Excel (.xlsx, .xlsm, .xltx, .xltm) or CSV (.csv)
- **Data Sheet**: For Excel files, the sheet must be named `PSGC`
- **Structure**: Hierarchical PSGC data with a header row, then these columns (any further columns are ignored):
  - Column 0: 10-digit PSGC Code
  - Column 1: Name/Description

## Installation

Install the command, which is named `go-psgc-parser`:

```bash
go install github.com/knvillena/go-psgc-parser@latest
```

Or build it from a clone of this repository as `psgc-parser`:

```bash
go build -o psgc-parser .
```

The examples below use `./psgc-parser`. If you installed it, type `go-psgc-parser` instead.

## Usage

```
psgc-parser [options] <input_file>
```

Use one of `-regions`, `-provinces`, `-cities`, `-barangays`, `-all`, or `-clean` per run. Options can come before or after the input file. Files are written to an `output/` directory inside the current directory.

### Direct Extraction from Raw PSGC Data

Extract geographic levels directly from raw hierarchical PSGC Excel files without cleaning. This assumes that the data file has a sheet called 'PSGC':

```bash
# Extract all levels directly from Excel file
./psgc-parser -all PSGC-3Q-2025-Publication-Datafile.xlsx

# Extract individual levels directly from Excel file
./psgc-parser -regions PSGC-3Q-2025-Publication-Datafile.xlsx
./psgc-parser -provinces PSGC-3Q-2025-Publication-Datafile.xlsx
./psgc-parser -cities PSGC-3Q-2025-Publication-Datafile.xlsx
./psgc-parser -barangays PSGC-3Q-2025-Publication-Datafile.xlsx
```

### Clean Raw PSGC Data

Convert hierarchical PSGC data to a flattened format, one row per barangay with its region, province, and city/municipality:

```bash
# Clean Excel file (reads from "PSGC" sheet)
# Output: output/psgc_cleaned_go.csv
./psgc-parser -clean PSGC-3Q-2025-Publication-Datafile.xlsx

# Clean CSV file
./psgc-parser -clean raw_psgc_data.csv

# Specify output file
./psgc-parser -clean input.xlsx -output my_cleaned.csv
```

### Extract Geographic Levels

The input can be the raw datafile (Excel or CSV) or a file made by `-clean`.

#### Extract All Levels at Once
```bash
# Directly from Excel file (preferred)
./psgc-parser -all PSGC-3Q-2025-Publication-Datafile.xlsx

# Or from cleaned CSV data
./psgc-parser -all output/psgc_cleaned_go.csv
```
This creates in the `output/` directory:
- `output/rregion.csv` - Regions
- `output/rprovince.csv` - Provinces
- `output/rcity.csv` - Cities/Municipalities
- `output/rbarangay.csv` - Barangays

#### Extract Individual Levels

| Option | Default output |
| --- | --- |
| `-regions` | `output/rregion.csv` |
| `-provinces` | `output/rprovince.csv` |
| `-cities` | `output/rcity.csv` |
| `-barangays` | `output/rbarangay.csv` |

```bash
# From Excel file (recommended)
./psgc-parser -provinces PSGC-3Q-2025-Publication-Datafile.xlsx

# Or from cleaned CSV data
./psgc-parser -provinces output/psgc_cleaned_go.csv
```

### Custom Output Files

Single-level extraction and `-clean` accept a custom output file:

```bash
./psgc-parser -regions PSGC-3Q-2025-Publication-Datafile.xlsx -output regions.csv
```

`-all` ignores `-output` and always writes the four files above.

## Output Formats

Each output file is **sorted in ascending order by its own code**.

### Regions (output/rregion.csv)
```csv
regcode,regname
01,Region I (Ilocos Region)
02,Region II (Cagayan Valley)
03,Region III (Central Luzon)
...
```

### Provinces (output/rprovince.csv)
```csv
regcode,provcode,provname
01,01028,Ilocos Norte
01,01029,Ilocos Sur
01,01033,La Union
...
```

### Cities (output/rcity.csv)
```csv
regcode,provcode,citycode,cityname
01,01028,0102801,Adams
01,01028,0102802,Bacarra
01,01028,0102803,Badoc
...
```

### Barangays (output/rbarangay.csv)
```csv
regcode,provcode,citycode,barangaycode,barangayname
01,01028,0102801,0102801001,Adams
01,01028,0102802,0102802001,Bani
01,01028,0102802,0102802002,Buyon
...
```

## PSGC Code Structure

```
10 digits: RRPPPCCBBB
- RR: Region Code (2 digits)
- PPP: Province Code (3 digits)
- CC: City/Municipality Code (2 digits)
- BBB: Barangay Code (3 digits)
```

## Workflow Examples

### Direct Extraction (Recommended)
Extract all geographic levels directly from the Excel file in one step:

```bash
./psgc-parser -all PSGC-3Q-2025-Publication-Datafile.xlsx
```

**Files generated:**
- `output/rregion.csv` - 18 regions
- `output/rprovince.csv` - 118 province-level entries
- `output/rcity.csv` - 1,655 cities, municipalities, and Manila districts
- `output/rbarangay.csv` - 42,011 barangays

### Two-Step Workflow
If you also need the cleaned data:

1. **Clean raw PSGC data:**
   ```bash
   ./psgc-parser -clean PSGC-3Q-2025-Publication-Datafile.xlsx
   ```

2. **Extract all geographic levels:**
   ```bash
   ./psgc-parser -all output/psgc_cleaned_go.csv
   ```

This writes `output/psgc_cleaned_go.csv` and the same four files as direct extraction.

## Notes

### Places Without a Province

Every city and barangay has a province code, so these places appear in the provinces file as their own entry. The counts are for the 3rd quarter 2025 datafile.

| Entries | Kind |
| --- | --- |
| 82 | Provinces |
| 33 | Highly urbanized cities, including the City of Manila |
| 1 | Pateros |
| 2 | City of Isabela (Not a Province) and Special Geographic Area |

- A highly urbanized city also appears in the cities file under its own code, for example `14,14303,1430300,City of Baguio`.
- The City of Manila is the exception: its 14 districts, such as Tondo I/II, are listed as cities under it, and Manila itself is not in the cities file.
- Pateros has no province-level row in the datafile, so the parser uses the first five digits of its code, `13817` (`1381700000` in the cleaned file).

### Codes

- All codes are zero-padded to ensure 10-digit format
- Province codes include region codes (RRPPP format)
- City codes include region + province codes (RRPPPCC format)
- Barangay codes are the full 10-digit PSGC codes
- Full code is RRPPPCCBBB

## Data Source

`PSGC-3Q-2025-Publication-Datafile.xlsx` is published by the Philippine Statistics Authority (PSA) and included here with acknowledgement of the PSA as the source. The PSA distributes the PSGC without warranty. The [MIT License](LICENSE) covers the code, not the datafile.
