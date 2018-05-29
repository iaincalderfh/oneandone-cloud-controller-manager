package oneandone

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"fmt"
	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"k8s.io/apimachinery/pkg/types"
	"errors"
)

func serverFromNode(node *v1.Node, client *oneandone.API) (*oneandone.Server, error) {
	nodeInstanceId, ok := node.Labels[oneandoneNodeInstanceIdLabel]
	if !ok {
		return nil, fmt.Errorf("serverFromNode: Node does not have the instanceId label: %s", oneandoneNodeInstanceIdLabel)
	}

	servers, err := client.ListServers(0, 0, "", nodeInstanceId, "id,name")
	if err != nil {
		return nil, fmt.Errorf("serverFromNode: Error looking up servers for matching: %s", nodeInstanceId)
	}

	for _, server := range servers {
		if nodeInstanceId == server.Name {
			fullServer, err := client.GetServer(server.Id)
			if err != nil {
				return nil, fmt.Errorf("serverFromNode: Error looking up server by id: %s", server.Id)
			}
			return fullServer, nil
		}
	}

	return nil, fmt.Errorf("serverFromNode: Could not find server with name: %s", nodeInstanceId)
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

	var privateIP string
	for _, ip := range node.Status.Addresses {
		if ip.Type == v1.NodeInternalIP {
			privateIP = ip.Address
			break
		}
	}

	if privateIP == "" {
		return nil, fmt.Errorf("could not get private ip: %v", )
	}
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: privateIP})

	publicIP  := server.Ips[0].Ip
	if publicIP == "" {
		return nil, fmt.Errorf("nodeAddresses: could not get public ip: %v", publicIP)
	}
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: publicIP})

	return addresses, nil
}