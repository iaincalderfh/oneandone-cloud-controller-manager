package oneandone

import (
	"errors"
	"fmt"

	"net"

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

	// Get Server by node public IP
	externalIP, err := getExternalIP(node)
	if err == nil {
		return nil, err
	}

	return serverFromExternalIP(string(externalIP), client)
}

func serverFromName(serverName string, client *oneandone.API) (*oneandone.Server, error) {
	servers, err := client.ListServers(0, 0, "", serverName, "id,name")
	if err != nil {
		return nil, fmt.Errorf("serverFromName: Error looking up servers for matching: %s", serverName)
	}

	for _, server := range servers {
		if serverName == server.Name {
			fullServer, err := client.GetServer(server.Id)
			if err != nil {
				return nil, fmt.Errorf("serverFromName: Error looking up server by id: %s", server.Id)
			}
			return fullServer, nil
		}
	}

	return nil, fmt.Errorf("serverFromName: Error Server with name: %s NOT_FOUND", serverName)
}

func serverFromExternalIP(externalIP string, client *oneandone.API) (*oneandone.Server, error) {
	servers, err := client.ListServers(0, 0, "", externalIP, "id,name,ips")
	if err != nil {
		return nil, fmt.Errorf("serverFromExternalIP: Error looking up servers for matching: %s", externalIP)
	}

	for _, server := range servers {
		for _, ip := range server.Ips {
			if externalIP == ip.Ip {
				fullServer, err := client.GetServer(server.Id)
				if err != nil {
					return nil, fmt.Errorf("serverFromExternalIP: Error looking up server by ip: %s", server.Id)
				}
				return fullServer, nil
			}
		}
	}

	return nil, fmt.Errorf("serverFromExternalIP: Error Server with ip: %s NOT_FOUND", externalIP)
}

func nodeFromName(nodeName types.NodeName, kubeClient v12.NodeInterface) (*v1.Node, error) {
	if kubeClient == nil {
		return nil, errors.New("nodeFromName: kubeClient is not instantiated")
	}

	node, err := kubeClient.Get(string(nodeName), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("nodeFromName: kubeClient error getting node by name: %s, error:%s", nodeName, err)
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
func nodeAddresses(server *oneandone.Server, node *v1.Node) ([]v1.NodeAddress, error) {
	var addresses []v1.NodeAddress
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeHostName, Address: node.Name})

	// Private IP Addresses
	for _, privateNetwork := range server.PrivateNets {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: privateNetwork.ServerIP})
	}

	// Public IP Addresses
	for _, serverIP := range server.Ips {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: serverIP.Ip})
	}

	return addresses, nil
}

func getExternalIP(node *v1.Node) (net.IP, error) {
	for _, ip := range node.Status.Addresses {
		if ip.Type == v1.NodeExternalIP {
			return net.ParseIP(ip.Address), nil
		}
	}

	return nil, fmt.Errorf("getExternalIP: Could not find an external ip for node: %s", node.Name)
}
