## Deploying CFCR with a Cloud Provider

Kubernetes exposes the concept of a [Cloud Provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
which interfaces with the IAAS to provision TCP Load Balancers, Nodes, Networking Routes, and Persistent Volumes.

CFCR can configure Kubernetes with a Cloud Provider through the following methods.
In each example it is assumed that you already have access to a BOSH Director.

### GCP

1. Create a service account and IAM profiles for your master and worker nodes.
   [This Terraform script](https://github.com/cloudfoundry/bosh-bootloader/blob/master/plan-patches/cfcr-gcp/terraform/cfcr_iam_override.tf)
   can be used as a reference guide.

   **Note: The service accounts should be used per CFCR deployment (NOT per director)**

1. Save the service account email addresses into a vars file that will be used to create the cloud-config

    ```bash
    $ export deployment_name="your deployment name"

    $ cat ${deployment_name}-cc-vars.yml

    cfcr_master_service_account_address: <master-service-account-email>
    cfcr_worker_service_account_address: <worker-service-account-email>
    deployment_name: <deployment-name>
    ```

1. Add a cloud config for the deployment with BOSH [generic configs](https://bosh.io/docs/configs/)
   ```bash
   $ export KD="path to kubo-deployment repo"

   $ bosh update-config --name ${deployment_name} \
      ${KD}/manifests/cloud-config/iaas/gcp/use-vm-extensions.yml \
      --type cloud \
      --vars-file ${deployment_name}-cc-vars.yml
   ```

1. Deploy CFCR

    ```bash
    $ bosh deploy -d ${deployment_name} \
    ${KD}/manifests/cfcr.yml \
    -o ${KD}/manifests/ops-files/iaas/gcp/cloud-provider.yml \
    -o ${KD}/manifests/ops-files/use-vm-extensions.yml \
    -o ${KD}/manifests/ops-files/rename.yml \
    -v deployment_name=${deployment_name}
    ```

1. To test that the cloud provider has been configured correctly, create a simple nginx deployment with an external load balancer

    ```bash
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/nginx-lb.yml

    # wait for the nginx service to get an external-ip
    $ kubectl get services

    $ export external_ip=$(kubectl get service/nginx -o jsonpath={.status.loadBalancer.ingress[0].ip})
    $ curl http://${external_ip}:80
    ```

### AWS

1. Create IAM profiles for your master and worker nodes.
   [This Terraform script](https://github.com/cloudfoundry/bosh-bootloader/blob/master/plan-patches/cfcr-aws/terraform/cfcr_iam_override.tf)
   can be used as a reference guide.

   **Note: The profiles should be used per CFCR deployment (NOT per director)**

1. Save the profile into a vars file that will be used to create the cloud-config

    ```bash
    $ export deployment_name="your deployment name"

    $ cat ${deployment_name}-cc-vars.yml

    master_iam_instance_profile: <master-iam-profile-name>
    worker_iam_instance_profile: <worker-iam-profile-name>
    cfcr_master_target_pool: <list-of-elbs-for-master>
    kubernetes_cluster_tag: <tag-for-k8s-cluster-components>
    deployment_name: <deployment-name>
    ```

1. Add a cloud config for the deployment with BOSH [generic configs](https://bosh.io/docs/configs/)
    ```bash
    $ export KD="path to kubo-deployment repo"

    $ bosh update-config --name ${deployment_name} \
    ${KD}/manifests/cloud-config/iaas/aws/use-vm-extensions.yml \
    --type cloud \
    --vars-file ${deployment_name}-cc-vars.yml
    ```

1. Deploy CFCR

    ```bash
    $ bosh deploy -d ${deployment_name} \
    ${KD}/manifests/cfcr.yml \
    -o ${KD}/manifests/ops-files/iaas/aws/cloud-provider.yml \
    -o ${KD}/manifests/ops-files/use-vm-extensions.yml \
    -o ${KD}/manifests/ops-files/iaas/aws/lb.yml \
    -o ${KD}/manifests/ops-files/rename.yml \
    -v deployment_name=${deployment_name}
    ```

1. To test that the cloud provider has been configured correctly, create a simple nginx deployment with an external load balancer

    ```bash
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/nginx-lb.yml

    # wait for the nginx service to get an external-ip
    $ kubectl get services

    $ export external_ip=$(kubectl get service/nginx -o jsonpath={.status.loadBalancer.ingress[0].hostname})
    $ curl http://${external_ip}:80
    ```

### vSphere

1. Create creds that will be used by the cloud-manager-controller in master nodes to talk to the vSphere api.
    Check [vSphere Cloud Provider documentation](https://vmware.github.io/vsphere-storage-for-kubernetes/documentation/vcp-roles.html) for list of privileges.

    **Note**: vSphere workers do not need any credentials.

1. Save the information to connect to the vCenter into deployment vars file.

    ```bash
    $ export deployment_name="your deployment name"

    $ cat ${deployment_name}-vars.yml
    vcenter_master_user: <user>
    vcenter_master_password: <password>
    vcenter_ip: <vsphere.server>
    vcenter_dc: <vsphere.datacenter>
    vcenter_ds: <vsphere.datastore>
    vcenter_vms: <vsphere.vm_folder>
    director_uuid: <BOSH director UUID>
    deployment_name: <deployment-name>
    ```

    See for mode details at [spec](../jobs/cloud-provider/spec) and at [vSphere
    Cloud Provider documentation](https://vmware.github.io/vsphere-storage-for-kubernetes/documentation/overview.html)
    You can get the BOSH director UUID from the output of `bosh env` command.

1. Add (if does not exist) a vSphere specific cloud config  with BOSH [generic configs](https://bosh.io/docs/configs/)

    ```bash
    cat << EOF > cfcr-cc-vm_extension-vsphere.yml
    vm_extensions:
    - cloud_properties:
        vmx_options:
          disk.enableUUID: "1"
      name: enable-disk-UUID
    EOF
    ```

    ```bash
    $ bosh update-config --name cfcr-cc-vm_extension-vsphere \
    cfcr-cc-vm_extension-vsphere.yml \
    --type cloud
    ```


1. Deploy CFCR

    ```bash
    cat << EOF > use-vm-extensions-vsphere-only.yml
    - type: replace
      path: /instance_groups/name=worker/vm_extensions?/-
      value: enable-disk-UUID
    EOF
    ```

    ```bash
    $ export KD="path to kubo-deployment repo"

    $ bosh deploy -d ${deployment_name} \
    ${KD}/manifests/cfcr.yml \
    -o ${KD}/manifests/ops-files/iaas/vsphere/cloud-provider.yml \
    -o use-vm-extensions-vsphere-only.yml \
    -o ${KD}/manifests/ops-files/rename.yml \
    -v deployment_name=${deployment_name} \
    --vars-file ${deployment_name}-vars.yml
    ```

   **NOTE:** If the *vSphere api* is behind a proxy create the following ops file and add it when deploying with `-o add-proxy.yml`
    ```bash
    cat << EOF > add-proxy.yml
    - type: replace
    path: /instance_groups/name=master/jobs/name=kube-controller-manager/properties/http_proxy?
    value: ((http_proxy))

    - type: replace
    path: /instance_groups/name=master/jobs/name=kube-controller-manager/properties/https_proxy?
    value: ((https_proxy))

    - type: replace
    path: /instance_groups/name=master/jobs/name=kube-controller-manager/properties/no_proxy?
    value: ((no_proxy))
    EOF
    ```

   **NOTE:** If *everything* is behind a proxy add the following ops file when
   deploying with `-o {KD}/kubo-deployment/manifests/ops-files/add-proxy.yml`

1. To test that the cloud provider has been configured correctly, create a simple a workload with persistence volume

    ```bash
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/storage-class-vsphere.yml
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/persistent-volume-claim.yml

    # wait for the volume to be attached
    $ kubectl describe pvc ci-claim

    # Type    Reason                 Age   From                         Message
    #  ----    ------                 ----  ----                         -------
    # Normal  ProvisioningSucceeded  31s   persistentvolume-controller  Successfully provisioned volume ...
    #
    ```

   **NOTE:** vSphere cloud-provider does not support service of type LoadBalancer

### Openstack

1. Save the information to connect to the OpenStack into deployment vars file.

    ```bash
    $ export deployment_name="your deployment name"

    $ cat ${deployment_name}-vars.yml
    auth_url: < authentication URL >
    openstack_domain: < domain >
    openstack_project_id: < tenant id >
    region: < region >
    openstack_username: < admin username >
    openstack_password: < admin user password >
    ```

    See for mode details at [spec](../jobs/cloud-provider/spec).
1. Deploy CFCR

   ```bash
   $ export KD="path to kubo-deployment repo"

   $ bosh deploy -d ${deployment_name} \
   ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/openstack/cloud-provider.yml \
   -o ${KD}/manifests/ops-files/rename.yml \
   -v deployment_name=${deployment_name} \
   --vars-file ${deployment_name}-vars.yml
   ```

1. To test that the cloud provider has been configured correctly, create a simple a workload with persistence volume

    ```bash
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/storage-class-openstack.yml
    $ kubectl apply -f https://github.com/cloudfoundry-incubator/kubo-ci/raw/master/specs/persistent-volume-claim.yml

    # wait for the volume to be attached
    $ kubectl describe pvc ci-claim

    # Type    Reason                 Age   From                         Message
    #  ----    ------                 ----  ----                         -------
    # Normal  ProvisioningSucceeded  31s   persistentvolume-controller  Successfully provisioned volume ...
    #
    ```
