# oneandone-cloud-controller-manager

## Setup and Installation

### Authentication and Configuration

Create a Kubernetes secret with the following command:

```bash
$ kubectl create secret generic oneandone-cloud-controller-manager \
     -n kube-system                                           \
     --from-file=cloud-provider.yaml=manifests/cloud-provider-example.yaml
```

Note that you must ensure the secret contains the key `cloud-provider.yaml`
rather than the name of the file on disk.