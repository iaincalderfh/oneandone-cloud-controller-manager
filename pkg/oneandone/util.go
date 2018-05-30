package oneandone

import (
	"errors"
	"fmt"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func serverFromNode(node *v1.Node, client *oneandone.API) (*oneandone.Server, error) {
	nodeInstanceID, ok := node.Labels[oneandoneNodeInstanceIDLabel]
	if !ok {
		return nil, fmt.Errorf("serverFromNode: Node does not have the instanceId label: %s", oneandoneNodeInstanceIDLabel)
	}

	servers, err := client.ListServers(0, 0, "", nodeInstanceID, "id,name")
	if err != nil {
		return nil, fmt.Errorf("serverFromNode: Error looking up servers for matching: %s", nodeInstanceID)
	}

	for _, server := range servers {
		if nodeInstanceID == server.Name {
			fullServer, err := client.GetServer(server.Id)
			if err != nil {
				return nil, fmt.Errorf("serverFromNode: Error looking up server by id: %s", server.Id)
			}
			return fullServer, nil
		}
	}

	return nil, fmt.Errorf("serverFromNode: Could not find server with name: %s", nodeInstanceID)
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

	if privateIP != "" {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: privateIP})
	} else {
		glog.V(1).Infof("nodeAddresses: could not get private ip for node: %v", node.Name)
	}

	publicIP := server.Ips[0].Ip
	if publicIP != "" {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: publicIP})
	} else {
		glog.V(1).Infof("nodeAddresses: could not get public ip for node: %v", node.Name)
	}

	return addresses, nil
}
