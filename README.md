# kubo-release

## What is this?
A BOSH release for [Kubernetes](http://kubernetes.io). You might also be interested in:
* [kubo meta](https://www.github.com/pivotal-cf-experimental/kubo-meta) - deployment scripts, manifests, and  the home of all our documentation
* [kubo service adapter release](https://www.github.com/pivotal-cf-experimental/kubo-service-adapter-release) -
  a [service adapter](https://docs.pivotal.io/on-demand-service-broker) for creating Kubernetes clusters on-demand via the Cloud Foundry API 


## Upgrading Kubernetes

### Maintaining offline support

This release can be deployed without external internet access. This is acomplished by loading any required containers into the docker engine of the worker nodes (in [post-start](./jobs/kubelet/templates/bin/post-start.erb)). To maintain this support ensure any updates that depend on new container are added as blobs. See the [download_container_images](./script/download_container_images) script to automate the fetch and add of new images.
