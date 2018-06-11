# Kubernetes Cloud Control Manager for 1and1

oneandone-cloud-controller-manager is the Kubernetes cloud controller manager implementation for 1and1. Read more about cloud controller managers [here](https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/). Running oneandone-cloud-controller-manager allows you to leverage the cloud provider features offered by 1and1 on your kubernetes clusters.

**WARNING**: this project is a work in progress.  Still TODO are:

 - better test coverage
 - automated end-to-end tests
 - Kubernetes compatibility definitions

## Setup and Installation

### Preparing Your Cluster

Node names or public IP addresses must match servers in your cloud panel.  If node names do not match, use one of these options:

- Add the label `stackpoint.io/instance_id` to **all nodes** in your cluster.  The label's value must be the server name.
- Set the `--hostname-override` flag on the `kubelet` on **all nodes** in your cluster.

Your cluster must be configured to use an external cloud-provider.

This involves:

- Setting the `--cloud-provider=external` flag on the `kubelet` on **all nodes** in your cluster.
- Setting the `--cloud-provider=external` flag on the `kube-controller-manager` in your Kubernetes control plane.

**WARNING**: setting `--cloud-provider=external` will taint all nodes in a cluster with `node.cloudprovider.kubernetes.io/uninitialized`.  It is the responsibility of a cloud controller manager to untaint those nodes once it has finished initializing them. This means that most pods will be left unschedulable until the cloud controller manager is running.

**Depending on how kube-proxy is run you _may_ need the following:**

- Ensure that `kube-proxy` tolerates the uninitialised cloud taint. The
  following should appear in the `kube-proxy` pod yaml:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```
If your cluster was created using `kubeadm` >= v1.7.2 this toleration will
already be applied. See [kubernetes/kubernetes#49017][5] for details.

**If you are running flannel, ensure that kube-flannel tolerates the uninitialised cloud taint:**
- The following should appear in the `kube-flannel` daemonset:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```



Remember to restart any components that you have reconfigured before continuing.

### Authentication and Configuration

The 1&1 Cloud Controller Manager requires a cloud panel API token and the datacenter code stored in the following environment variables:

- ONEANDONE_API_KEY
- ONEANDONE_INSTANCE_REGION

The default manifest is configured to set these environment variables from a secret named `oneandone`:

kubectl -n kube-system create secret generic oneandone --from-literal=token=`<TOKEN>`
--from-literal=credentials-datacenter=GB

### Installation - with RBAC

`kubectl apply -f manifests/oneandone-ccm-rbac.yaml`

`kubectl apply -f manifests/oneandone-ccm.yaml`

### Installation - without RBAC
`kubectl apply -f manifests/oneandone-ccm.yaml`