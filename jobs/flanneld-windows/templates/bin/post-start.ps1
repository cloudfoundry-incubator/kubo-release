$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

# add back GCP / AWS metadata server
<% if p("backend-type") == "win-overlay" %>
  <% name = "vxlan0" %>
<% else %>
  <% name = "cbr0" %>
<% end %>
while (!(Get-HNSNetwork | ? Name -Eq "<%= name %>"))
{
    Write-Host "Waiting for overlay network to be enabled"
    Start-Sleep -sec 1
}
route /p add 169.254.169.254 mask 255.255.255.255 0.0.0.0
