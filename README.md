# Cloud Foundry Container Runtime
A [BOSH](http://bosh.io/) release for [Kubernetes](http://kubernetes.io).  Formerly named **kubo**.

- **Slack**: #cfcr on https://slack.cloudfoundry.org
- **Pivotal Tracker**: https://www.pivotaltracker.com/n/projects/2093412

## Using CFCR Docs (DEPRECATED)

We are no longer supporting the following documentation for deploying BOSH and CFCR:
* https://docs-cfcr.cfapps.io

The [deploy_bosh](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/bin/deploy_bosh)
and [deploy_k8s](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/bin/deploy_k8s)
scripts in the `kubo-deployment` repository are now deprecated.

## Prerequisites
- A BOSH Director configured with UAA, Credhub, and BOSH DNS.
- [kubo-release](https://github.com/cloudfoundry-incubator/kubo-release)
- [kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment)
- Cloud Config with 
  - `vm_types` named `minimal`, `small`, and `small-highmem` (See [cf-deployment](https://github.com/cloudfoundry/cf-deployment) for reference)
  - `network` named `default`
  - There are three availability zones `azs`, and they are named `z1`,`z2`,`z3`
  - note: the cloud-config properties can be customized by applying ops-files. See `manifests/ops-files` for some examples

## Deploying CFCR

#### Single Master
1. Upload the appropriate stemcell to the director. You can determine the version and type of the stemcell with  
	```
	bosh int ~/workspace/kubo-deployment/manifests/cfcr.yml --path /stemcells
	```
1. Copy the link to the [latest version of kubo-release tarball](https://github.com/cloudfoundry-incubator/kubo-release/releases/latest) and upload it to the director  
1. Deploy
	```
	cd kubo-deployment

	bosh deploy -d cfcr manifests/cfcr.yml \
	  -o manifests/ops-files/misc/single-master.yml
	```
	If you have a **BOSH Lite** environment, run
	```
	cd kubo-deployment

	bosh deploy -d cfcr manifests/cfcr.yml \
	  -o manifests/ops-files/misc/single-master.yml \
	  -o manifests/ops-files/iaas/virtualbox/bosh-lite.yml
	```
1. Add kubernetes system components
	```
	bosh -d cfcr run-errand apply-specs
	```
1. Run the following to confirm the cluster is operational
	```
	bosh -d cfcr run-errand smoke-tests
	```
## Accessing the CFCR Cluster (kubectl)

### Without Load Balancer
1. Login to the Credhub Server that stores the cluster's credentials:
	```
	credhub login
	```
1. Find the IP address of one master node by running 
	```
	bosh -d cfcr vms
	```
1. Configure the `kubeconfig` for your `kubectl` client:
	```
	cd kubo-deployment

	./bin/set_kubeconfig <DIRECTOR_NAME>/cfcr https://<master_node_IP_address>:8443
	```
