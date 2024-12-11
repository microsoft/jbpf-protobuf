#!/bin/bash
echo "Running gofmt pre-commit hook..."

# Define the gofmt executable
GOFMT_BIN=gofmt

# Check if gofmt is installed
if ! command -v ${GOFMT_BIN} > /dev/null; then
  echo "Error: gofmt is not installed."
  exit 0
fi

# Get a list of all staged Go files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.go$')

# Exit if there are no files to format
if [ -z "$STAGED_FILES" ]; then
  echo "No Go files staged for commit."
  exit 0
fi

# If the env AUTO_FORMAT is set to 1, format the files
if [ "$AUTO_FORMAT" = "1" ]; then
  echo "Auto-formatting staged files..."
  for FILE in ${STAGED_FILES}; do
    ${GOFMT_BIN} -w "${FILE}"
    git add "${FILE}" # Re-add formatted files to staging
  done
  exit 0
fi

# Check the formatting of each staged file
FORMAT_ERRORS=0
for FILE in ${STAGED_FILES}; do
  echo "Checking formatting of: ${FILE} ..."  
  if ! ${GOFMT_BIN} -l "${FILE}"; then
    echo "Error: ${FILE} is not formatted correctly."
    FORMAT_ERRORS=1
  fi
done

# If there were formatting errors, fail the commit
if [ $FORMAT_ERRORS -ne 0 ]; then
  echo "Error: Some files are not formatted correctly. Please run gofmt -w and stage the changes."
  exit 1
fi

# Exit successfully
exit 0
