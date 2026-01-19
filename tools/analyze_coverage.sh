#!/bin/bash

# Script to analyze profile.cov and report code coverage for spectrocloud package
# Usage: ./analyze_coverage.sh [profile.cov]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default coverage file
COVERAGE_FILE="${1:-profile.cov}"

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo -e "${RED}Error: Coverage file '$COVERAGE_FILE' not found${NC}" >&2
    echo "Usage: $0 [profile.cov]" >&2
    echo "Looking for profile.cov in current directory or common locations..." >&2
    exit 1
fi

echo -e "${BLUE}=== Spectrocloud Code Coverage Analysis ===${NC}\n"
echo "Coverage file: $COVERAGE_FILE"
echo ""

# Check if go tool cover is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: 'go' command not found. Please install Go.${NC}" >&2
    exit 1
fi

# Create temporary files for processing
TEMP_COVERAGE=$(mktemp)
TEMP_PROCESSED=$(mktemp)
TEMP_FILE_COVERAGE=$(mktemp)
trap "rm -f $TEMP_COVERAGE $TEMP_PROCESSED $TEMP_FILE_COVERAGE" EXIT

# Filter coverage file to only include spectrocloud files
# Coverage file format: github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/...
set +e
grep "/spectrocloud/" "$COVERAGE_FILE" > "$TEMP_COVERAGE" 2>/dev/null
GREP_EXIT=$?
set -e

if [ $GREP_EXIT -ne 0 ] || [ ! -s "$TEMP_COVERAGE" ]; then
    echo -e "${YELLOW}Warning: No spectrocloud files found in coverage file${NC}"
    echo "Checking if coverage file contains any data..."
    if [ ! -s "$COVERAGE_FILE" ]; then
        echo -e "${RED}Error: Coverage file is empty${NC}" >&2
        exit 1
    fi
    # Try using the whole file
    cp "$COVERAGE_FILE" "$TEMP_COVERAGE"
fi

# Count total lines in filtered coverage
TOTAL_LINES=$(wc -l < "$TEMP_COVERAGE" | tr -d ' ')

if [ "$TOTAL_LINES" -eq 0 ]; then
    echo -e "${RED}Error: No coverage data found for spectrocloud package${NC}" >&2
    exit 1
fi

echo -e "${GREEN}Found $TOTAL_LINES lines of coverage data for spectrocloud${NC}\n"

# Calculate total coverage directly from coverage file (faster than go tool cover -func)
# Handle duplicate blocks by using unique keys
echo -e "${BLUE}--- Overall Coverage Statistics ---${NC}"
TOTAL_STATS=$(awk '
/^mode:/ { next }
/\/spectrocloud\// {
    # Use the full line as unique key to avoid duplicate counting
    key = $1
    count = $2
    covered = $3
    
    # Track unique blocks - covered is execution count, not covered statement count
    # If covered > 0, the block is covered (regardless of how many times)
    if (key in block_seen) {
        # If block was covered in any test run, mark it as covered
        if (covered > 0) {
            block_seen[key] = 1
        }
    } else {
        block_seen[key] = (covered > 0 ? 1 : 0)
        block_count[key] = count
    }
}
END {
    total = 0
    covered_blocks = 0
    for (key in block_seen) {
        total += block_count[key]
        if (block_seen[key] > 0) {
            covered_blocks += block_count[key]
        }
    }
    if (total > 0) {
        coverage = (covered_blocks / total) * 100
        printf "%.1f%%", coverage
    } else {
        printf "0.0%%"
    }
}' "$TEMP_COVERAGE")
TOTAL_COVERAGE="$TOTAL_STATS"

# Also get function-level output for reference (but don't wait for it if it's slow)
COVERAGE_OUTPUT=$(go tool cover -func="$TEMP_COVERAGE" 2>&1 || echo "")

# Aggregate function-level coverage to file-level coverage
# Parse the coverage file directly to get accurate statement counts and coverage
# Coverage file format: mode: set
# file:start.end count covered
awk '
/^mode:/ { next }
/\/spectrocloud\// {
    key = $1
    file = $1
    sub(/:[0-9]+\.[0-9]+,.*/, "", file)
    gsub(/.*\/spectrocloud\//, "", file)
    count = $2
    covered = $3
    
    # covered is execution count, not covered statement count
    # If covered > 0, the block is covered
    if (key in block_seen) {
        # If block was covered in any test run, mark it as covered
        if (covered > 0) {
            block_seen[key] = 1
        }
    } else {
        block_seen[key] = (covered > 0 ? 1 : 0)
        block_count[key] = count
        block_file[key] = file
    }
}
END {
    for (key in block_seen) {
        file = block_file[key]
        count = block_count[key]
        is_covered = block_seen[key]
        file_statements[file] += count
        if (is_covered > 0) {
            file_covered[file] += count
        }
    }
    for (file in file_statements) {
        total = file_statements[file]
        covered = file_covered[file]
        if (total > 0) {
            coverage_pct = (covered / total) * 100
            printf "%-70s %12d %11.1f%%\n", file, total, coverage_pct
        } else {
            printf "%-70s %12d %11s\n", file, total, "0.0%"
        }
    }
}' "$TEMP_COVERAGE" | sort -k3 -rn > "$TEMP_FILE_COVERAGE"

# Check if file aggregation worked
if [ ! -s "$TEMP_FILE_COVERAGE" ]; then
    echo -e "${YELLOW}Warning: Could not aggregate file-level coverage${NC}"
    echo "Trying alternative method..."
    # Fallback: use go tool cover output directly
    go tool cover -func="$TEMP_COVERAGE" 2>&1 | grep "/spectrocloud/" | \
        awk '{file=$1; sub(/:[0-9]+:.*/, "", file); gsub(/.*\/spectrocloud\//, "", file); coverage=$NF; gsub(/%/, "", coverage); if(file in file_coverage) {file_coverage[file] = (file_coverage[file] + coverage) / 2} else {file_coverage[file] = coverage; file_count[file] = 1}} END {for(f in file_coverage) printf "%-70s %12d %11.1f%%\n", f, file_count[f], file_coverage[f]}' | \
        sort -k3 -rn > "$TEMP_FILE_COVERAGE"
fi

if [ -z "$TOTAL_COVERAGE" ]; then
    echo -e "${YELLOW}Warning: Could not extract total coverage. Showing full output:${NC}"
    echo "$COVERAGE_OUTPUT"
else
    # Display total coverage with color coding
    COVERAGE_NUM=$(echo "$TOTAL_COVERAGE" | sed 's/%//' | awk '{print int($1)}')
    if [ "$COVERAGE_NUM" -ge 80 ] 2>/dev/null; then
        COLOR=$GREEN
    elif [ "$COVERAGE_NUM" -ge 60 ] 2>/dev/null; then
        COLOR=$YELLOW
    else
        COLOR=$RED
    fi
    
    echo -e "Total Coverage: ${COLOR}${TOTAL_COVERAGE}${NC}"
    echo ""
    
    # Show per-file coverage summary (top 10 lowest and highest)
    echo -e "${BLUE}--- Coverage by File (Lowest 10) ---${NC}"
    sort -k3 -n "$TEMP_FILE_COVERAGE" | head -10 | \
        awk '{printf "%-70s %12s %12s\n", $1, $2, $3}'
    
    echo ""
    echo -e "${BLUE}--- Coverage by File (Highest 10) ---${NC}"
    head -10 "$TEMP_FILE_COVERAGE" | \
        awk '{printf "%-70s %12s %12s\n", $1, $2, $3}'
    
    echo ""
    echo -e "${BLUE}--- Detailed Statistics ---${NC}"
    
    # Count files by coverage ranges
    LOW_COUNT=$(awk '{cov=substr($3,1,length($3)-1); if (cov+0 < 50) print}' "$TEMP_FILE_COVERAGE" | wc -l | tr -d ' ')
    MEDIUM_COUNT=$(awk '{cov=substr($3,1,length($3)-1); if (cov+0 >= 50 && cov+0 < 80) print}' "$TEMP_FILE_COVERAGE" | wc -l | tr -d ' ')
    HIGH_COUNT=$(awk '{cov=substr($3,1,length($3)-1); if (cov+0 >= 80) print}' "$TEMP_FILE_COVERAGE" | wc -l | tr -d ' ')
    TOTAL_FILES=$(wc -l < "$TEMP_FILE_COVERAGE" | tr -d ' ')
    
    echo "Files with coverage < 50%:  ${RED}$LOW_COUNT${NC}"
    echo "Files with coverage 50-79%: ${YELLOW}$MEDIUM_COUNT${NC}"
    echo "Files with coverage >= 80%: ${GREEN}$HIGH_COUNT${NC}"
    echo "Total files analyzed:        $TOTAL_FILES"
fi

echo ""
echo -e "${BLUE}--- Coverage by File (All Files) ---${NC}"
echo ""
printf "%-70s %12s %12s\n" "File" "Statements" "Coverage"
echo "--------------------------------------------------------------------------------"

# Display each file with color-coded coverage percentage
while IFS= read -r line; do
    if [ -n "$line" ]; then
        file=$(echo "$line" | awk '{print $1}')
        statements=$(echo "$line" | awk '{print $2}')
        coverage=$(echo "$line" | awk '{print $3}')
        cov_num=$(echo "$coverage" | sed 's/%//' | awk '{print int($1)}')
        
        if [ "$cov_num" -ge 80 ] 2>/dev/null; then
            printf "${GREEN}%-70s %12s %12s${NC}\n" "$file" "$statements" "$coverage"
        elif [ "$cov_num" -ge 60 ] 2>/dev/null; then
            printf "${YELLOW}%-70s %12s %12s${NC}\n" "$file" "$statements" "$coverage"
        elif [ "$cov_num" -ge 50 ] 2>/dev/null; then
            printf "${YELLOW}%-70s %12s %12s${NC}\n" "$file" "$statements" "$coverage"
        else
            printf "${RED}%-70s %12s %12s${NC}\n" "$file" "$statements" "$coverage"
        fi
    fi
done < "$TEMP_FILE_COVERAGE"

echo ""
echo -e "${GREEN}Analysis complete!${NC}"
