## Using Proxy Properties in the Release

To enable the use of a proxy the following [ops-file](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/manifests/ops-files/add-proxy.yml) can be applied.

### Common values for `NO_PROXY`

Webhooks / Extending Kubernetes API:

* .svc
* .svc.cluster.local
* localhost
* <service_network_ip_range>

By default the `service_network_ip_range` is: 10.100.200.0/24

Typical Registries:

* registry-1.docker.io
* auth.docker.io
* production.cloudflare.docker.com
* gcr.io
* storage.googleapis.com





