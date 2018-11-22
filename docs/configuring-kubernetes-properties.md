## Configuring Kubernetes Properties in the Release
In [v0.24.0](https://github.com/cloudfoundry-incubator/kubo-release/releases/tag/v0.24.0) of kubo-release we exposed pretty much all of the Kubernetes configuration options. Our intent is to allow users to configure Kubernetes however they like at the BOSH release level but continue to provide our opinion via the manifest and ops-files in [kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment/tree/master/manifests). 

This change will affect the following Kubernetes components in kubo-release
- kube-apiserver
- kube-controller-manager
- kube-proxy
- kube-scheduler
- kubelet
- cloud-provider

### How to Configure
Instead of detailing every configuration option in Kubernetes and updating them with each version change, we're going to allow users to pass through the configuration options documented on the official Kubernetes doc site in a couple of ways:
- [`k8s-args` Option](#k8s-args-job-property)
- [Configuration File Option](#config-file-option)
- [Cloud Provider Configuration](#cloud-provider-configuration)

#### `k8s-args` option
Snippet of the k8s documentation for a kube-apiserver flag option:
```
--audit-log-format string Default: "json"

Format of saved audits. "legacy" indicates 1-line text format for each event. "json" indicates structured json format. Known formats are legacy,json.
```

To configure this property in kubo-release, the manifest should look like: 
```
jobs:
 - name: kube-apiserver
    properties:
      k8s-args:
        audit-log-format: legacy
       
```

#### Configuration File Option
One thing to note is that Kubernetes is in the process of deprecating some of the command line flag options in favor of moving all configurations to a file specified by `--config`. For instance, `kube-proxy` is a component that will emit deprecation warning logs for most configuration options provided via command line flags instead of through `--config`. For components that have already started moving to this process, we've provided an additional job property `[jobname]-configuration`. Here's an example:

Snippet of the k8s documentation for a kube-proxy flag option:
```
--feature-gates mapStringBool
A set of key=value pairs that describe feature gates for alpha/experimental features. Options are:

APIListChunking=true|false (BETA - default=true)
APIResponseCompression=true|false (ALPHA - default=false)
AllAlpha=true|false (ALPHA - default=false)
...
```

bosh manifest for kubo-release:
```
jobs:
 - name: kube-proxy
    properties:
      kube-proxy-configuration:
        apiVersion: kubeproxy.config.k8s.io/v1alpha1
        kind: KubeProxyConfiguration
        clientConnection:
          kubeconfig: /var/vcap/jobs/kube-proxy/config/kubeconfig
        featureGates:
          APIResponseCompression: true
          AllAlpha: true
          ...
```

#### Cloud Provider Configuration
The cloud provider differs in a couple ways from the other Kubernetes components described in this documentation. It's an optional job that we provide [example ops-files](https://github.com/cloudfoundry-incubator/kubo-deployment/tree/master/manifests/ops-files/iaas) for each of the supported IaaSes. We recommend using these ops-files for configuring the cloud provider. 

One thing to be aware of is the format of the cloud provider cloud-config's file. The format may differ for different cloud providers. For instance, Azure's cloud provider will result in a `yaml` file while most of the other supported cloud providers will be `ini`. 

### Unsupported Config Options
Our intention is to have the least possible amount of logic and opinion baked into the kubo-release. However there are a few properties we have set to make life easier. Although configuring the following properties may not result in templating errors, we cannot gurantee the behavior of duplicating these flags. Configure at your own risk!

Note `cloud-config` and `config` are still configurable as described above. However we currently hardcode the file path the configuration options are templated to. 

- kube-apiserver
  - `apiserver-count`
  - `cloud-provider` type
  - `cloud-config` path
  - `etcd-servers`
- kube-controller-manager
  - `cloud-config` path
  - `cloud-provider` type
- kube-proxy
  - `config` path
- kube-scheduler
  - `config` path
- kubelet
  - `cloud-config` path
  - `cloud-provider` type
  - `hostname-override`
  - `node-labels`
  - `config` path
