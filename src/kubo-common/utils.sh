#!/bin/bash -exu

detect_cloud_config() {
  if [ -e /var/vcap/jobs/cloud-provider/bin/cloud-provider_utils ]; then
    source /var/vcap/jobs/cloud-provider/bin/cloud-provider_utils
    set_cloud_provider
    cloud_config="/var/vcap/jobs/cloud-provider/config/cloud-provider.ini"
  else
    cloud_config=""
    cloud_provider=""
  fi
}

check_and_chmod_product_serial() {
  local cloud_provider = $1
  if [ "$cloud_provider" != "" ] && [ -f /sys/class/dmi/id/product_serial ]; then
    chmod a+r /sys/class/dmi/id/product_serial
  fi
}
