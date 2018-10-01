$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

C:\var\vcap\packages\kubelet-windows-poststart\bin\poststart.exe `
--kubeconfig=C:\var\vcap\jobs\kubelet-windows\config\kubeconfig `
--nodeip=<%= spec.ip %>
