$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

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
