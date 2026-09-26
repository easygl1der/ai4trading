#!/usr/bin/env bash
# Download arXiv PDFs listed in research/papers/manifest.json into research/papers/arxiv/.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MANIFEST="${ROOT}/research/papers/manifest.json"
OUT_DIR="${ROOT}/research/papers/arxiv"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

is_valid_pdf() {
  [[ -f "$1" ]] && [[ "$(file -b --mime-type "$1")" == "application/pdf" ]]
}

mapfile -t IDS < <(jq -r '.arxiv[].id' "$MANIFEST")

for id in "${IDS[@]}"; do
  dest="${OUT_DIR}/${id}.pdf"
  if is_valid_pdf "$dest"; then
    echo "skip ${id} (exists)"
    continue
  fi
  rm -f "$dest"
  url="https://arxiv.org/pdf/${id}.pdf"
  echo "fetch ${id}"
  if ! curl -fsSL --retry 3 --retry-delay 2 -o "$dest" "$url"; then
    echo "failed ${id}" >&2
    rm -f "$dest"
    continue
  fi
  # arXiv sometimes returns HTML error pages with a .pdf filename
  if ! is_valid_pdf "$dest"; then
    echo "not a PDF: ${id}" >&2
    rm -f "$dest"
  fi
done

echo "done. PDF count: $(find "$OUT_DIR" -maxdepth 1 -name '*.pdf' | wc -l)"
