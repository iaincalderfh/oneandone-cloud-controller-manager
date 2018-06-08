package oneandone

import (
	"errors"
	"fmt"

	"github.com/leroyshirtoFH/oneandone-cloudserver-sdk-go"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func serverFromNode(node *v1.Node, client *oneandone.API) (*oneandone.Server, error) {
	// Get Server by name identified by oneandoneNodeInstanceIDLabel
	nodeInstanceID, ok := node.Labels[oneandoneNodeInstanceIDLabel]
	if ok {
		return serverFromName(nodeInstanceID, client)
	}

	// Get Server by name identified by node name
	server, err := serverFromName(node.Name, client)
	if err == nil && server != nil {
		return server, nil
	}

	// Get Server by node public IPs
	externalAddresses := getNodeAddressesByType(node, v1.NodeExternalIP)
	server, err = serverFromExternalIPs(externalAddresses, client)
	if err == nil && server != nil {
		return server, nil
	}

	return nil, fmt.Errorf("serverFromNode: Unable to find server for node '%s'", node.Name)
}

func serverFromName(serverName string, client *oneandone.API) (*oneandone.Server, error) {
	servers, err := client.ListServers(0, 0, "", serverName, "id,name")
	if err != nil {
		return nil, fmt.Errorf("serverFromName: Error looking up servers matching '%s': %s", serverName, err)
	}

	for _, server := range servers {
		if serverName == server.Name {
			fullServer, err := client.GetServer(server.Id)
			if err != nil {
				return nil, fmt.Errorf("serverFromName: Error looking up server by id '%s': %s", server.Id, err)
			}
			return fullServer, nil
		}
	}

	return nil, fmt.Errorf("serverFromName: Server with name '%s' NOT_FOUND", serverName)
}

func serverFromExternalIPs(externalAddresses []v1.NodeAddress, client *oneandone.API) (*oneandone.Server, error) {
	for _, address := range externalAddresses {
		servers, err := client.ListServers(0, 0, "", address.Address, "id,ips")
		if err != nil {
			return nil, fmt.Errorf("serverFromExternalIPs: Error looking up servers matching '%s': %s", address.Address, err)
		}

		for _, server := range servers {
			for _, ip := range server.Ips {
				if address.Address == ip.Ip {
					fullServer, err := client.GetServer(server.Id)
					if err != nil {
						return nil, fmt.Errorf("serverFromExternalIPs: Error looking up server by ip '%s': %s", server.Id, err)
					}
					return fullServer, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("serverFromExternalIPs: No Servers matching public addresses %v", externalAddresses)
}

func nodeFromName(nodeName types.NodeName, kubeClient v12.NodeInterface) (*v1.Node, error) {
	if kubeClient == nil {
		return nil, errors.New("nodeFromName: kubeClient is not instantiated")
	}

	node, err := kubeClient.Get(string(nodeName), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("nodeFromName: kubeClient error getting node by name '%s': %s", nodeName, err)
	}

	return node, nil
}

func serverFromNodeName(nodeName types.NodeName, client *oneandone.API, kubeClient v12.NodeInterface) (*oneandone.Server, error) {
	node, err := nodeFromName(nodeName, kubeClient)
	if err != nil {
		return nil, err
	}

	server, err := serverFromNode(node, client)
	if err != nil {
		return nil, err
	}

	return server, nil
}

// nodeAddresses returns a []v1.NodeAddress from server.
func nodeAddresses(server *oneandone.Server, node *v1.Node) []v1.NodeAddress {
	addresses := []v1.NodeAddress{v1.NodeAddress{Type: v1.NodeHostName, Address: node.Name}}

	// Private IP Addresses
	for _, privateNetwork := range server.PrivateNets {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: privateNetwork.ServerIP})
	}

	// Public IP Addresses
	for _, serverIP := range server.Ips {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: serverIP.Ip})
	}

	return addresses
}

func getNodeAddressesByType(node *v1.Node, addressType v1.NodeAddressType) []v1.NodeAddress {
	var addresses []v1.NodeAddress
	for _, address := range node.Status.Addresses {
		if address.Type == addressType {
			addresses = append(addresses, address)
		}
	}

	return addresses
}
