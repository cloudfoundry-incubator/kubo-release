# kubo-release
A [BOSH](http://bosh.io/) release for [Kubernetes](http://kubernetes.io).

**Slack**: #kubo on https://slack.cloudfoundry.org

## Deploying Kubo
Deployment instructions, scripts, and manifests are located in [kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment).

## Developing Kubo

### Upgrading Kubernetes

#### Maintaining offline support
This release can be deployed without external internet access. This is accomplished by loading any required containers into the docker engine of the worker nodes (in [post-start](./jobs/kubelet/templates/bin/post-start.erb)). To maintain this support ensure any updates that depend on new container are added as blobs. See the [download_container_images](./script/download_container_images) script to automate the fetch and add of new images.
