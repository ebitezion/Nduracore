#!/usr/bin/env bash
set -euo pipefail

# Only check tracked Go sources inside this repo.
# Explicitly skip .cache paths and ignore missing files that can linger in the index.
unformatted=""
while IFS= read -r file; do
  [[ -f "$file" ]] || continue
  result=$(gofmt -l "$file")
  if [[ -n "$result" ]]; then
    unformatted+="$result"$'\n'
  fi
done < <(git ls-files '*.go' | grep -v '^\.cache/' || true)

if [[ -n "$unformatted" ]]; then
  echo "gofmt found unformatted files:"
  printf "%s" "$unformatted"
  exit 1
fi

echo "gofmt check passed"
