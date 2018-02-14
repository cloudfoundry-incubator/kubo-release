#!/bin/bash -exu

check_and_chmod_product_serial() {
  local cloud_provider=$1
  if [ "$cloud_provider" != "" ] && [ -f /sys/class/dmi/id/product_serial ]; then
    chmod a+r /sys/class/dmi/id/product_serial
  fi
}
