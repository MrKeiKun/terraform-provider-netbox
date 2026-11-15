#!/bin/bash
# Fails when NetBox has a release that the CI matrix does not cover: a release that is newer than the newest
# tested version of its minor release, or a whole minor release that the matrix does not have. Releases that are
# older than the newest tested one of their minor release are not a problem, and neither are releases older
# than the oldest supported one.

set -euo pipefail

workflow=.github/workflows/ci.yml
oldest=v4.4.0

tested="$(grep -o "all='\[[^]]*\]'" "$workflow" | grep -o 'v[0-9][0-9.]*' || true)"
if [ -z "$tested" ]; then
    echo "::error::Could not read the list of NetBox versions from $workflow"
    exit 1
fi

released="$(gh api --paginate repos/netbox-community/netbox/tags --jq '.[].name' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V)"
if [ -z "$released" ]; then
    echo "::error::Could not read the NetBox releases"
    exit 1
fi

missing=""
for tag in $released; do
    # only releases from the oldest supported one on
    if ! printf '%s\n%s\n' "$oldest" "$tag" | sort -V -C; then
        continue
    fi
    minor="${tag%.*}"
    newest="$(printf '%s\n' "$tested" | grep -F "$minor." | grep -E "^${minor//./\.}\.[0-9]+$" | sort -V | tail -1 || true)"
    if [ -z "$newest" ] || ! printf '%s\n%s\n' "$tag" "$newest" | sort -V -C; then
        missing="$missing $tag"
    fi
done

if [ -n "$missing" ]; then
    for tag in $missing; do
        echo "::error::NetBox $tag is released, but the matrix in $workflow does not test it or a newer version of its minor release"
    done
    echo "Add the versions to the 'all' list (and the 'edges' list when it is the newest of its minor release) in $workflow."
    exit 1
fi
echo "Every NetBox release from $oldest on is covered by the matrix."
