#!/bin/bash
# Fetches the OpenAPI schema of one NetBox version per minor release into the given directory, with the
# same docker setup as the acceptance tests. The files are the input of
#
#   go run ./tools/netboxdrift -write-versions internal/fwprovider/versions_generated.go \
#     -versions "4.4.0=DIR/schema-v4.4.10.json,4.5.0=DIR/schema-v4.5.10.json,4.6.0=DIR/schema-v4.6.10.json,4.7.0=DIR/schema-v4.7.5.json"
#
# Usage: tools/netboxdrift/fetch-schemas.sh DIR [VERSION...]
set -euo pipefail

dir=${1:?usage: fetch-schemas.sh DIR [VERSION...]}
shift
versions=("$@")
if [ ${#versions[@]} -eq 0 ]; then
    versions=(v4.4.10 v4.5.10 v4.6.10 v4.7.5)
fi
mkdir -p "$dir"

export SUPERUSER_API_TOKEN=0123456789abcdef0123456789abcdef01234567
export SUPERUSER_API_KEY=abcdef012345
export API_TOKEN_PEPPER_1=0123456789abcdefghij0123456789abcdefghij0123456789abcdefghij
compose=(docker compose -p netbox-schema -f docker/docker-compose.yml)

for v in "${versions[@]}"; do
    export NETBOX_VERSION=$v
    "${compose[@]}" up -d
    for _ in $(seq 1 60); do
        [ "$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8001/login/)" = 200 ] && break
        sleep 10
    done
    curl -sf -o "$dir/schema-$v.json" "http://localhost:8001/api/schema/?format=json"
    "${compose[@]}" down -v
done
