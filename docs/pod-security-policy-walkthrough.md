
# PodSecurityPolicy in CFCR

Since CFCR v0.24 you can now enable the PodSecurityPolicy admission controller. 

This will detail how to enable PodSecurityPolicy, with an example workflow to apply and bind an example policy. More information can be found in the kubernetes [documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy).

## Deploy with PodSecurityPolicy
When deploying your cluster using BOSH, ensure the following operations file is applied:

[master/manifests/ops-files/enable-podsecuritypolicy.yml](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/manifests/ops-files/enable-podsecuritypolicy.yml)

This will add the `PodSecurityPolicy` entry to the `enable-admission-plugins` property in the manifest for the `kube-apiserver`.

:exclamation: **Warning:** You must ensure appropriate policies are applied and bound to exsting workloads before enabling the PodSecurityPolicy in admission controllers. Any existing workloads will **stop working** if they are not bound to a policy, or the policy does not have the right permissions. :exclamation:

## Apply a policy and binding

### 1. Apply the policy. 

This example is using the [restricted example](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/policy/restricted-psp.yaml) from kubernetes [documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).
```
$ cat restricted-psp.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default'
    apparmor.security.beta.kubernetes.io/allowedProfileNames: 'runtime/default'
    seccomp.security.alpha.kubernetes.io/defaultProfileName:  'docker/default'
    apparmor.security.beta.kubernetes.io/defaultProfileName:  'runtime/default'
spec:
  privileged: false
  # Required to prevent escalations to root.
  allowPrivilegeEscalation: false
  # This is redundant with non-root + disallow privilege escalation,
  # but we can provide it for defense in depth.
  requiredDropCapabilities:
    - ALL
  # Allow core volume types.
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    # Assume that persistentVolumes set up by the cluster admin are safe to use.
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    # Require the container to run without root privileges.
    rule: 'MustRunAsNonRoot'
  seLinux:
    # This policy assumes the nodes are using AppArmor rather than SELinux.
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  readOnlyRootFilesystem: false
  ```
`$ kubectl apply -f restricted-psp.yaml`

### 2. Apply the role binding

```
$ cat restricted-rolebinding.yml
 kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: psp:restricted
roleRef:
  kind: Role
  name: psp:restricted
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  apiGroup: rbac.authorization.k8s.io
  # system:authenticated captures all authenticated users
  name: system:authenticated
```
`$ kubectl apply -f restricted-rolebinding.yml -n default
`

_This will bind the policy `psp:restricted` to all authenticated users in the namespace, in this case we are applying this to the 'default' namespace_

You can find more information on Roles and RoleBindings in the kubernetes [RBAC documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

## 3. Deploy the workload. 
Here we use a demo pod with security context configuration, originally sourced [here](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/).

```
$ cat security-context-demo.yml
apiVersion: v1
kind: Pod
metadata:
  name: security-context-demo
spec:
  securityContext:
    runAsUser: 1000
    fsGroup: 2000
  volumes:
  - name: sec-ctx-vol
    emptyDir: {}
  containers:
  - name: sec-ctx-demo
    image: gcr.io/google-samples/node-hello:1.0
    volumeMounts:
    - name: sec-ctx-vol
      mountPath: /data/demo
    securityContext:
      allowPrivilegeEscalation: false
```
`$ kubectl apply -f security-context-demo.yml -n default`

The `restricted` policy applied above will allow this pod to be deployed as the policy allows for `runAsUser`, and it is applied to the appropriate namespace.
