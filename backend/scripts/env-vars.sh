#!/bin/bash

# 1. Ask for the file path
echo "Please enter the path to the YAML file:"
read -r file_path

# 2. Check if file exists
if [[ ! -f "$file_path" ]]; then
    echo "Error: File '$file_path' not found."
    exit 1
fi

echo -e "\n--- Extracted Environment Variables ---\n"

# 3. Use sed to extract and reformat
# Logic:
# - Find lines containing ${VAR:VAL}
# - Remove any characters before and after the ${...}
# - Replace the : separator with =
# - Remove the ${ and } brackets
# - Sort and uniq to remove duplicates (common in your example)

sed -n 's/.*${\([^}]*\)}.*/\1/p' "$file_path" | \
sed 's/:/=/ ' | \
sort | uniq

echo -e "\n---------------------------------------"