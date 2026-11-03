#!/usr/bin/env bash
# Fetch OFL-licensed fonts (Inter, Victor Mono, JetBrains Mono) from the
# Fontsource npm packages via jsDelivr, together with their licence texts.
# Run once with `bash scripts/fetch-fonts.sh`, then commit the results.
#
# PP Fraktion Mono is personal-use / commercial — drop .woff2 onto the
# deploy volume or local machine; do not commit it here.
set -euo pipefail
cd "$(dirname "$0")/.."

dest=internal/ui/static/fonts
cdn=https://cdn.jsdelivr.net/npm/@fontsource
cdn_var=https://cdn.jsdelivr.net/npm/@fontsource-variable
mkdir -p "$dest"

# Fontsource versions are pinned so the files (and their licence) are stable.
inter=5.2.5
inter_var=5.2.5
victor=5.3.0
jetbrains=5.3.0

get() { # url outfile
  curl -fsSL "$1" -o "$dest/$2"
  echo "  $2"
}

echo "Inter Variable…"
get "$cdn_var/inter@$inter_var/files/inter-latin-opsz-normal.woff2" "inter-latin-opsz-normal.woff2"
# Fallback wght axis if opsz package layout differs.
if [[ ! -s "$dest/inter-latin-opsz-normal.woff2" ]]; then
  get "$cdn_var/inter@$inter_var/files/inter-latin-wght-normal.woff2" "inter-latin-opsz-normal.woff2"
fi
get "$cdn_var/inter@$inter_var/LICENSE" OFL-inter-variable.txt

echo "Inter static weights…"
for w in 400 500 600 700; do
  get "$cdn/inter@$inter/files/inter-latin-$w-normal.woff2" "inter-latin-$w-normal.woff2"
done
get "$cdn/inter@$inter/LICENSE" OFL-inter.txt

echo "Victor Mono…"
for w in 400 700; do
  get "$cdn/victor-mono@$victor/files/victor-mono-latin-$w-normal.woff2" "victor-mono-latin-$w-normal.woff2"
done
get "$cdn/victor-mono@$victor/LICENSE" OFL-victor-mono.txt

echo "JetBrains Mono…"
for w in 400 700; do
  get "$cdn/jetbrains-mono@$jetbrains/files/jetbrains-mono-latin-$w-normal.woff2" "jetbrains-mono-latin-$w-normal.woff2"
done
get "$cdn/jetbrains-mono@$jetbrains/LICENSE" OFL-jetbrains-mono.txt

echo "Done → $dest"
