#!/usr/bin/env bash

set -eu

: "${IAAS:-?}"

source .github/tasks/set-bosh-env source_file

VM=""
if [ ${IAAS} == "gcp" ]; then
  IAAS="google"
  VM="kvm"
elif [ ${IAAS} == "aws" ]; then
  VM="xen-hvm"
elif [ ${IAAS} == "vsphere" ]; then
  VM="esxi"
elif [ ${IAAS} == "azure" ]; then
  VM="hyperv"
elif [ ${IAAS} == "openstack" ]; then
  VM="kvm"
fi

stemcell_version="$(bosh int --path=/stemcells/0/version manifests/cfcr.yml)"
stemcell_line="$(bosh int --path=/stemcells/0/os manifests/cfcr.yml)"

bosh upload-stemcell --name="bosh-${IAAS}-${VM}-${stemcell_line}-go_agent" --version="${stemcell_version}" "https://s3.amazonaws.com/bosh-core-stemcells/${stemcell_version}/bosh-stemcell-${stemcell_version}-${IAAS}-${VM}-${stemcell_line}-go_agent.tgz"