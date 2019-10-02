$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$AddedFolder= "C:\var\vcap\packages\docker-windows\docker\"

$OldPath=(Get-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH).Path

if (-not $OldPath.Contains($AddedFolder)) {
  $NewPath=$OldPath+';'+$AddedFolder
  Set-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH -Value $newPath
}

C:\var\vcap\packages\docker-windows\docker\dockerd --register-service

Start-Service Docker
