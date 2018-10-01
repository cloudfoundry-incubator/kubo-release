$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

# add back GCP / AWS metadata server
route /p add 169.254.169.254 mask 255.255.255.255 0.0.0.0
