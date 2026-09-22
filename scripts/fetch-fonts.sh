#!/usr/bin/env bash
# Fetch the OFL-licensed fallback fonts (Victor Mono and JetBrains Mono) from
# the Fontsource npm packages via jsDelivr, together with their licence texts.
# Run once with `bash scripts/fetch-fonts.sh`, then commit the results.
set -euo pipefail
cd "$(dirname "$0")/.."

dest=internal/ui/static/fonts
cdn=https://cdn.jsdelivr.net/npm/@fontsource
mkdir -p "$dest"

# Fontsource versions are pinned so the files (and their licence) are stable.
victor=5.3.0
jetbrains=5.3.0

get() { # url outfile
  curl -fsSL "$1" -o "$dest/$2"
  echo "  $2"
}

for w in 400 700; do
  get "$cdn/victor-mono@$victor/files/victor-mono-latin-$w-normal.woff2" "victor-mono-latin-$w-normal.woff2"
  get "$cdn/jetbrains-mono@$jetbrains/files/jetbrains-mono-latin-$w-normal.woff2" "jetbrains-mono-latin-$w-normal.woff2"
done

# OFL condition: every copy of the fonts carries the copyright and licence.
get "$cdn/victor-mono@$victor/LICENSE" OFL-victor-mono.txt
get "$cdn/jetbrains-mono@$jetbrains/LICENSE" OFL-jetbrains-mono.txt
