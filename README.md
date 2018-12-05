# Cloud Foundry Container Runtime
A [BOSH](http://bosh.io/) release for [Kubernetes](http://kubernetes.io).  Formerly named **kubo**.

- **Slack**: #cfcr on https://slack.cloudfoundry.org
- **Pivotal Tracker**: https://www.pivotaltracker.com/n/projects/2093412

## Prerequisites
- A BOSH Director configured with UAA, Credhub, and [BOSH DNS runtime config](https://raw.githubusercontent.com/cloudfoundry/bosh-deployment/master/runtime-configs/dns.yml). We recommend using [BOSH Bootloader](https://github.com/cloudfoundry/bosh-bootloader) for this.
- [Latest kubo-deployment tarball](https://github.com/cloudfoundry-incubator/kubo-deployment/releases/latest)
- Accessing the master:
  - **Single Master:** Set up a DNS name pointing to your master's IP address
  - **Multiple Masters:** A TCP load balancer for your master nodes.
    - Use a TCP load balancer configured to connect to the master nodes on port 8443.
    - Add healthchecks using either a TCP dial or HTTP by looking for a `200 OK` response from `/healthz`.
    - if you have used [BOSH Bootloader](https://github.com/cloudfoundry/bosh-bootloader) on GCP then you need to manually create a firewall rule.  Allow access to port TCP 8443 to VMs in your BBL network tagged `cfcr-master` from your load balancer's IP.
- Cloud Config with
  - `vm_types` named `minimal`, `small`, and `small-highmem` (See [cf-deployment](https://github.com/cloudfoundry/cf-deployment) for reference)
  - `network` named `default`
  - three availability zones `azs` named `z1`,`z2`,`z3`

  Note: the cloud-config properties can be customized by applying ops-files. See `manifests/ops-files` for some examples.
  
  If using loadbalancers then apply the `vm_extension` called `cfcr-master-loadbalancer` to the cloud-config to add the instances to your loadbalancers. See [BOSH documentation](https://bosh.io/docs/cloud-config/#vm-extensions) for information on how to configure loadbalancers.

#### Hardware Requirements
Kubernetes uses etcd as its datastore. The official infrastructure requirements and example configurations for the etcd cluster can be found [here](https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/hardware.md).

## Deploying CFCR

1. Upload the [latest Xenial stemcell](https://bosh.io/stemcells/#ubuntu-xenial) to the director.

1. Untar the kubo-deployment tarball and rename it `kubo-deployment`

1. Deploy

    ##### Option 1. Single Master

	```bash
	cd kubo-deployment

	bosh deploy -d cfcr manifests/cfcr.yml \
	  -o manifests/ops-files/misc/single-master.yml \
	  -o manifests/ops-files/add-hostname-to-master-certificate.yml \
	  -v api-hostname=[DNS-NAME]
	```

    ##### Option 2. Three Masters

	```bash
	cd kubo-deployment

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
### Configuring CFCR
Please check out our manifest and ops-files in kube-deployment for examples on how to configure kubo-release. Additionally, we have a [doc page](docs/configuring-kubernetes-properties.md) to describe how to configure Kubernetes components for the release.

### BOSH Lite
CFCR clusters on BOSH Lite are intended for development. We run the [deploy_cfcr_lite](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/bin/deploy_cfcr_lite) script to provision a cluster with the latest stemcell and master of kubo-release.  This requires that the cloned kubo-release repository can be found from `cd ../kubo-release` from within the kubo-deployment directory.

```
cd kubo-deployment
./bin/deploy_cfcr_lite
```
## Accessing the CFCR Cluster with kubectl

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
## Backup & Restore
We use [BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore) to perform backups and restores of the etcd node within a CFCR cluster, for both single and three master deployments. Our backup currently takes an etcd snapshot without interruptions to the cluster. However, for restore we take both the kube-apiserver and etcd offline to restore the cluster with the specified snapshot. Restore is a destructive operation that will completely overwrite any existing data on the cluster. For a closer look at the bbr scripts, check out:
- [cfcr-etcd-release](https://github.com/cloudfoundry-incubator/cfcr-etcd-release/tree/master/jobs/bbr-etcd)
- [kubo-release](https://github.com/cloudfoundry-incubator/kubo-release/tree/master/jobs/bbr-kube-apiserver)

To run the `bbr` cli against a CFCR cluster, follow the steps under "BOSH Deployment" on the BBR [documentation page](https://docs.cloudfoundry.org/bbr/#bosh-deployment).

## Monitoring

Follow the recommendations in [etcd's documentation](https://github.com/etcd-io/etcd/blob/master/Documentation/metrics.md) for monitoring etcd
metrics.

## Deprecations

### Deployment scripts and docs
CFCR had a set of scripts, including `deploy_bosh` and `deploy_k8s`, that were the primary mechanism we supported to deploy BOSH and Kubernetes clusters. We no longer support these and have removed the corresponding documentation from https://docs-cfcr.cfapps.io

The BOSH oriented method documented in this README.md is the supported method to deploy Kubernetes clusters with CFCR.

### Heapster
K8s 1.11 release kicked off the deprecation timeline for the Heapster component, see [here](https://github.com/kubernetes/heapster/blob/master/docs/deprecation.md) for more info. As a result, we're in the process of replacing Heapster with [Metrics Server](https://github.com/kubernetes-incubator/metrics-server) in the upcoming releases of kubo-release.
