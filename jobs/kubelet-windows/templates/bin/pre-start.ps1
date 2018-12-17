$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$env:PATH+=";C:\var\vcap\packages\docker\docker\;"

# Temporary workaround until we bring our own images
docker pull microsoft/nanoserver:1803
docker pull microsoft/windowsservercore:1803

if (!(docker images microsoft/nanoserver:latest -q))
{
    docker tag (docker images microsoft/nanoserver -q) microsoft/nanoserver
}

if (!(docker images microsoft/windowsservercore:latest -q))
{
    docker tag (docker images microsoft/windowsservercore -q) microsoft/windowsservercore
}

$infraPodImage=docker images kubeletwin/pause -q
if (!$infraPodImage)
{
    cd /var/vcap/jobs/kubelet-windows/config
    docker build -t kubeletwin/pause .
}

Set-NetFirewallProfile -Profile Domain, Public, Private -Enabled False

<% if_p('cloud-provider') do |cloud_provider| %>
  <% if cloud_provider == "vsphere" %>
# Override the hostname to work around
# vSphere cloud provider ignoring hostname override
# and kubernetes requiring all-lowercase node names
$ComputerName = (cat C:\var\vcap\bosh\settings.json | ConvertFrom-Json).agent_id
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" -name "Hostname" -value $ComputerName
  <% end %>
<% end %>

# Needed until https://github.com/kubernetes/kubernetes/pull/71147 is merged
mkdir -force /sys/class/dmi/id
(wmic csproduct get IdentifyingNumber).Split([Environment]::Newline)[2] | Out-File -Encoding ASCII /sys/class/dmi/id/product_serial
