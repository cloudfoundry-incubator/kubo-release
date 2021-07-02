<img src="/images/cfcr_white.png?raw=true" width="600" height="200">

A [BOSH](http://bosh.io/) release for [Kubernetes](http://kubernetes.io).  Formerly named **kubo**.

- **Slack**: #cfcr on https://slack.cloudfoundry.org
- **2021 Roadmap**: https://github.com/cloudfoundry-incubator/kubo-release/projects/3

## Build Status

| Iaas | Upgrades | Conformance | Integration | Turbulence | Istio |  PSPs |
| --- | --- | -- | -- | -- | -- | -- |
|GCP | [![GCP Upgrades](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_old-release_upgrade/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_old-release_upgrade) | [![GCP Conformance](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_vanilla_conformance/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_vanilla_conformance) | [![GCP Integration](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_vanilla_integration/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_vanilla_integration) | [![GCP Turbulence](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_vanilla_turbulence/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_vanilla_turbulence) | [![GCP Istio](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_vanilla_istio/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_vanilla_istio) | [![GCP PSPs](https://ci.kubo.sh/api/v1/teams/main/pipelines/gcp_pod-security-policy_integration/jobs/run-tests/badge)](https://ci.kubo.sh/teams/main/pipelines/gcp_pod-security-policy_integration) |

# Table of Contents
<!-- vscode-markdown-toc -->
* [Prerequisites](#Prerequisites)
  * [Hardware Requirements](#HardwareRequirements)
* [Deploying CFCR](#DeployingCFCR)
  * [Configuring CFCR](#ConfiguringCFCR)
  * [Using Proxy with CFCR](#ProxyWithCFCR)
* [Accessing the CFCR Cluster with kubectl](#AccessingtheCFCRClusterwithkubectl)
* [Backup & Restore](#BackupRestore)
* [Monitoring](#Monitoring)

<!-- vscode-markdown-toc-config
	numbering=true
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

##  <a name='Prerequisites'></a>Prerequisites
- A BOSH Director configured with UAA, Credhub, and [BOSH DNS runtime config](https://raw.githubusercontent.com/cloudfoundry/bosh-deployment/master/runtime-configs/dns.yml). We recommend using [BOSH Bootloader](https://github.com/cloudfoundry/bosh-bootloader) for this.
- [Latest kubo-deployment tarball](https://github.com/cloudfoundry-incubator/kubo-deployment/releases/latest)
- Accessing the master:
  - **Single Master:** Set up a DNS name pointing to your master's IP address
  - **Multiple Masters:** A TCP load balancer for your master nodes.
    - Use a TCP load balancer configured to connect to the master nodes on port 8443.
    - Add healthchecks using either a TCP dial or HTTPS by looking for a `200 OK` response from `/healthz`.
    - if you have used [BOSH Bootloader](https://github.com/cloudfoundry/bosh-bootloader) on GCP then you need to manually create a firewall rule.  Allow access to port TCP 8443 to VMs in your BBL network tagged `cfcr-master` from your load balancer's IP.
- Cloud Config with
  - `vm_types` named `minimal`, `small`, and `small-highmem` (See [cf-deployment](https://github.com/cloudfoundry/cf-deployment) for reference)
  - `network` named `default`
  - three availability zones `azs` named `z1`,`z2`,`z3`

  Note: the cloud-config properties can be customized by applying ops-files. See `manifests/ops-files` for some examples.
  
  If using loadbalancers then apply the `vm_extension` called `cfcr-master-loadbalancer` to the cloud-config to add the instances to your loadbalancers. See [BOSH documentation](https://bosh.io/docs/cloud-config/#vm-extensions) for information on how to configure loadbalancers.

####  <a name='HardwareRequirements'></a>Hardware Requirements
Kubernetes uses etcd as its datastore. The official infrastructure requirements and example configurations for the etcd cluster can be found [here](https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/hardware.md).

##  <a name='DeployingCFCR'></a>Deploying CFCR

1. Upload the [latest Bionic stemcell](https://bosh.io/stemcells/#ubuntu-bionic) to the director.

1. Untar the kubo-release tarball and rename it `kubo-release`

1. Deploy

    ##### Option 1. Single Master

	```bash
	cd kubo-release

	bosh deploy -d cfcr manifests/cfcr.yml \
	  -o manifests/ops-files/misc/single-master.yml \
	  -o manifests/ops-files/add-hostname-to-master-certificate.yml \
	  -v api-hostname=[DNS-NAME]
	```

    ##### Option 2. Three Masters

	```bash
	cd kubo-release

	bosh deploy -d cfcr manifests/cfcr.yml \
	  -o manifests/ops-files/add-vm-extensions-to-master.yml \
	  -o manifests/ops-files/add-hostname-to-master-certificate.yml \
	  -v api-hostname=[LOADBALANCER-ADDRESS]
	```

	*Note: Loadbalancer address should be the external address (hostname or IP) of the loadbalancer you have configured.*

   Check additional configurations, such as setting Kubernetes cloud provider, in [docs](./docs/cloud-provider.md).

1. Add Kubernetes system components

    ```bash
    bosh -d cfcr run-errand apply-specs
    ```

1. Run the following to confirm the cluster is operational

    ```bash
    bosh -d cfcr run-errand smoke-tests
    ```
###  <a name='ConfiguringCFCR'></a>Configuring CFCR
Please check out our manifest and ops-files in kube-deployment for examples on how to configure kubo-release.
Additionally, we have a [doc page](docs/configuring-kubernetes-properties.md) to describe how to configure Kubernetes components for the release.

CFCR can be deployed with Pod Security Policies. Check for more details in [the
doc](docs/pod-security-policy-walkthrough.md)

####  <a name='ProxyWithCFCR'></a>Configuring Proxy for CFCR
CFCR allows you to configure proxy for all components. Check [recommendations
for no proxy settings](docs/using-proxy.md) first.

##  <a name='AccessingtheCFCRClusterwithkubectl'></a>Accessing the CFCR Cluster with kubectl

1. Login to the Credhub Server that stores the cluster's credentials:
	```
	credhub login
	```
1. Find the director name by running
	```
	bosh env
	```
1. Configure the `kubeconfig` for your `kubectl` client:
	```
	cd kubo-deployment

	./bin/set_kubeconfig <DIRECTOR_NAME>/cfcr https://[DNS-NAME-OR-LOADBALANCER-ADDRESS]:8443
	```
##  <a name='BackupRestore'></a>Backup & Restore
We use [BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore) to perform backups and restores of the etcd node within a CFCR cluster, for both single and three master deployments. Our backup currently takes an etcd snapshot without interruptions to the cluster. However, for restore we take both the kube-apiserver and etcd offline to restore the cluster with the specified snapshot. Restore is a destructive operation that will completely overwrite any existing data on the cluster. For a closer look at the bbr scripts, check out:
- [cfcr-etcd-release](https://github.com/cloudfoundry-incubator/cfcr-etcd-release/tree/master/jobs/bbr-etcd)
- [kubo-release](https://github.com/cloudfoundry-incubator/kubo-release/tree/master/jobs/bbr-kube-apiserver)

To run the `bbr` cli against a CFCR cluster, follow the steps under "BOSH Deployment" on the BBR [documentation page](https://docs.cloudfoundry.org/bbr/#bosh-deployment).

##  <a name='Monitoring'></a>Monitoring

Follow the recommendations in [etcd's documentation](https://github.com/etcd-io/etcd/blob/master/Documentation/metrics.md) for monitoring etcd
metrics.
