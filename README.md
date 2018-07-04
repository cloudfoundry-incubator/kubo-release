# Cloud Foundry Container Runtime
A [BOSH](http://bosh.io/) release for [Kubernetes](http://kubernetes.io).  Formerly named **kubo**.

**Slack**: #cfcr on https://slack.cloudfoundry.org
**Pivotal Tracker**: https://www.pivotaltracker.com/n/projects/2093412

## Prerequisites
- A BOSH Director configured with UAA, Credhub, and BOSH DNS.
- [kubo-release](https://github.com/cloudfoundry-incubator/kubo-release)
- [kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment)

## Accessing the CFCR Cluster (kubectl)

### Without Load Balancer
1. Find the IP address of one master node e.g. `bosh -e ENV -d cfcr vms`
2. Login to the Credhub Server that stores the cluster's credentials:
  ```
  credhub login
  ```
3. Execute the [`./bin/set_kubeconfig` script](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/bin/set_kubeconfig) to configure the `kubeconfig` for use in your `kubectl` client:
  ```
  cd kubo-deployment

  $ ./bin/set_kubeconfig <ENV>/cfcr https://<master_node_IP_address>:8443
  ```
