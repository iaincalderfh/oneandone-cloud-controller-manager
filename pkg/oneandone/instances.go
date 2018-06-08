package oneandone

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"net/http"

	"github.com/leroyshirtoFH/oneandone-cloudserver-sdk-go"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// oneandoneNodeInstanceIdLabel is the label specifying the unique identifier of the node
	// or server name on the oneandone api.
	oneandoneNodeInstanceIDLabel = "stackpoint.io/instance_id"
)

type instances struct {
	client       *oneandone.API
	region       string
	kubeNodesAPI v12.NodeInterface
}

func newInstances(client *oneandone.API, region string) *instances {
	return &instances{client, region, nil}
}

// NodeAddresses implements cloudprovider.Instances.NodeAddresses.
// Returns: all the valid addresses of the server identified by nodeName.
// Only the public/private IPv4 addresses are considered for now.
func (i *instances) NodeAddresses(ctx context.Context, nodeName types.NodeName) ([]v1.NodeAddress, error) {
	node, err := nodeFromName(nodeName, i.kubeNodesAPI)
	if err != nil {
		return nil, err
	}

	server, err := serverFromNodeName(nodeName, i.client, i.kubeNodesAPI)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(server, node), nil
}

// NodeAddressesByProviderID implements cloudprovider.Instances.NodeAddressesByProviderID.
func (i *instances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	server, err := i.client.GetServer(id)
	if err != nil {
		return nil, err
	}

	node, err := nodeFromName(types.NodeName(server.Name), i.kubeNodesAPI)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(server, node), nil
}

// ExternalID implements cloudprovider.Instances.ExternalID.
// Returns: the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (i *instances) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return i.InstanceID(ctx, nodeName)
}

// InstanceID implements cloudprovider.Instances.InstanceID.
// Returns: the cloud provider ID of the node with the specified NodeName.
func (i *instances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	server, err := serverFromNodeName(nodeName, i.client, i.kubeNodesAPI)
	if err != nil {
		return "", err
	}

	return server.Id, nil
}

// InstanceType implements cloudprovider.Instances.InstanceType.
// Returns: the type of the specified instance.
func (i *instances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	server, err := serverFromNodeName(name, i.client, i.kubeNodesAPI)
	if err != nil {
		return "", err
	}

	return server.ServerType, nil
}

// InstanceTypeByProviderID implements cloudprovider.Instances.InstanceTypeByProviderID.
// Returns: the type of the specified instance.
func (i *instances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	server, err := i.client.GetServer(id)
	if err != nil {
		return "", err
	}

	return server.ServerType, nil
}

// AddSSHKeyToAllInstances implements cloudprovider.Instances.AddSSHKeyToAllInstances.
// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (i *instances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return errors.New("not implemented")
}

// CurrentNodeName implements cloudprovider.Instances.CurrentNodeName.
// Returns: the name of the node we are currently running on.
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (i *instances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// InstanceExistsByProviderID implements cloudprovider.Instances.InstanceExistsByProviderID.
// Returns: true if the instance for the given provider id still is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (i *instances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return false, err
	}

	_, err = i.client.GetServer(id)
	if err == nil {
		return true, nil
	}

	oneandoneErr, ok := err.(oneandone.ApiError)
	if !ok {
		return false, fmt.Errorf("unexpected error type from oneandone api: %T, msg: %v", err, err)
	}

	if oneandoneErr.HttpStatusCode() != http.StatusNotFound {
		return false, fmt.Errorf("error checking if instance exists: %v", err)
	}

	return false, nil
}

// serverIDFromProviderID returns a server's ID from providerID.
//
// The providerID spec should be retrievable from the
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
	if strings.TrimSuffix(split[0], ":") != providerName {
		return "", fmt.Errorf("provider name from providerID should be %s: %s", providerName, providerID)
	}

	return split[2], nil
}
