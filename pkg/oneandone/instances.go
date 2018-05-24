package oneandone

import (
	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/api/core/v1"
	"errors"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"fmt"
)

type instances struct {
	client *oneandone.API
	region string
	kubeNodesApi v12.NodeInterface
}

func newInstances(client *oneandone.API, region string) *instances {
	return &instances{client, region, nil}
}

// NodeAddresses returns all the valid addresses of the droplet identified by
// nodeName. Only the public/private IPv4 addresses are considered for now.
//
// When nodeName identifies more than one server, only the first will be
// considered.
func (i *instances) NodeAddresses(ctx context.Context, nodeName types.NodeName) ([]v1.NodeAddress, error) {
	server, err := serverFromNodeName(nodeName, i.client, i.kubeNodesApi)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(server)
}

func (i *instances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	server, err := i.serverByID(id)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(server)
}
// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (i *instances) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return i.InstanceID(ctx,  nodeName)
}
// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (i *instances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	server, err := serverFromNodeName(nodeName, i.client, i.kubeNodesApi)
	if err != nil {
		return "", err
	}

	return server.Id, nil
}
// InstanceType returns the type of the specified instance.
func (i *instances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	server, err := serverFromNodeName(name, i.client, i.kubeNodesApi)
	if err != nil {
		return "", err
	}

	return server.ServerType, nil
}
// InstanceTypeByProviderID returns the type of the specified instance.
func (i *instances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	server, err := i.serverByID(id)
	if err != nil {
		return "", err
	}

	return server.ServerType, nil
}
// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (i *instances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return errors.New("not implemented")
}
// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (i *instances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}
// InstanceExistsByProviderID returns true if the instance for the given provider id still is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (i *instances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	return false, errors.New("not implemented")
}

// serverByID returns a *oneandone.Server value for the sserver identified by id.
func (i *instances) serverByID(id string) (*oneandone.Server, error) {
	server, err := i.client.GetServer(id)
	if err != nil {
		return nil, err
	}

	return server, nil
}

// serverIDFromProviderID returns a server's ID from providerID.
//
// The providerID spec should be retrievable from the Kubernetes
// node object. The expected format is: oneandone://server-id
func serverIDFromProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", errors.New("providerID cannot be empty string")
	}

	split := strings.Split(providerID, "/")
	if len(split) != 3 {
		return "", fmt.Errorf("unexpected providerID format: %s, format should be: oneandone://F320EC38B4BF73F80C8AAA230334FB28", providerID)
	}

	// since split[0] is actually "oneandone:"
	if strings.TrimSuffix(split[0], ":") != ProviderName {
		return "", fmt.Errorf("provider name from providerID should be oneandone: %s", providerID)
	}

	return split[2], nil
}