#!/usr/bin/env bash

set -eu

: "${RELEASE_LIST:?}"
: "${STEMCELL_ALIAS:-default}"

pushd manifests
    STEMCELL_OS=$(bosh int cfcr.yml -o ops-files/windows/add-worker.yml --path /stemcells/alias=${STEMCELL_ALIAS}/os)
    STEMCELL_VERSION=$(bosh int cfcr.yml -o ops-files/windows/add-worker.yml --path /stemcells/alias=${STEMCELL_ALIAS}/version)

    export RELEASES=""
    for rel in $RELEASE_LIST; do
        release_url=$(bosh int cfcr.yml -o ops-files/non-precompiled-releases.yml -o ops-files/windows/add-worker.yml --path=/releases/name=$rel/url)
        release_version=$(bosh int cfcr.yml -o ops-files/non-precompiled-releases.yml -o ops-files/windows/add-worker.yml --path=/releases/name=$rel/version)
        RELEASES="$RELEASES- name: $rel\n  url: ${release_url}\n  version: ${release_version}\n"
    done
popd

cat > compilation-manifest/manifest.yml <<EOF
---
name: compilation-${STEMCELL_ALIAS}
releases:
$(echo -e "${RELEASES}")
stemcells:
- alias: default
  os: ${STEMCELL_OS}
  version: "${STEMCELL_VERSION}"
update:
  canaries: 1
  max_in_flight: 1
  canary_watch_time: 1000 - 90000
  update_watch_time: 1000 - 90000
instance_groups: []
EOF