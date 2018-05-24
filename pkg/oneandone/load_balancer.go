package oneandone

import (
	"context"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"fmt"
	"errors"
)

const (
	// oneandoneLoadbalancerProtocol is the annotation used to specify the default protocol
	// for oneandone load balancers. For ports specified in annDOTLSPorts, this protocol
	// is overwritten to https. Options are tcp, http and https. Defaults to tcp.
	oneandoneLoadbalancerProtocol = "service.beta.kubernetes.io/oneandone-loadbalancer-protocol"

	// defaultActiveTimeout is the number of seconds to wait for a load balancer to
	// reach the active state.
	defaultActiveTimeout = 90

	// oneandoneLoadbalancerMethod is the annotation specifying method the loadBalancer
	// should use. Options are ROUND_ROBIN and LEAST_CONNECTIONS. Defaults
	// to ROUND_ROBIN.
	oneandoneLoadbalancerMethod = "service.beta.kubernetes.io/oneandone-loadbalancer-method"

	// oneandoneNodeInstanceIdLabel is the label specifying the unique identifier of the node
	// or server name on the oneandone api.
	oneandoneNodeInstanceIdLabel = "stackpoint.io/instance_id"

	// defaultActiveCheckTick is the number of seconds between load balancer
	// status checks when waiting for activation.
	defaultActiveCheckTick = 5

	// statuses for oneandone load balancer
	lbStatusActive  = "ACTIVE"
)

var errLBNotFound = errors.New("loadbalancer not found")

// Compile-time check that loadBalancer implements cloudprovider. LoadBalancer
var _ cloudprovider.LoadBalancer = &loadBalancer{}

type loadBalancer struct {
	client            *oneandone.API
	region            string
	lbActiveTimeout   int
	lbActiveCheckTick int
}

type HealthCheck struct {
	CheckTest       string
	CheckInterval   int
	CheckPath       string
	CheckPathParser string
	Persistence     bool
	PersistenceTime int
}



// newLoadbalancers returns a cloudprovider.LoadBalancer whose concrete type is a *loadbalancer.
func newLoadbalancer(client *oneandone.API, region string) cloudprovider.LoadBalancer {
	return &loadBalancer{client, region, defaultActiveTimeout, defaultActiveCheckTick}
}

// GetLoadBalancer returns the *v1.LoadBalancerStatus of service.
//
// GetLoadBalancer will not modify service.
func (lb *loadBalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	glog.V(1).Infof("GetLoadBalancer: service=%s", service.Name)

	lbName := cloudprovider.GetLoadBalancerName(service)
	loadBalancer, err := lb.lbByName(ctx, lbName)
	if err != nil {
		if err == errLBNotFound {
			return nil, false, nil
		}

		return nil, false, err
	}

	if loadBalancer.State != lbStatusActive {
		err = lb.client.WaitForState(loadBalancer, lbStatusActive, 10, 60)
		if err != nil {
			return nil, true, fmt.Errorf("error waiting for load balancer to be active %v", err)
		}
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: loadBalancer.Ip,
			},
		},
	}, true, nil

	return nil, false, nil
}

// EnsureLoadBalancer ensures that the cluster is running a load balancer for
// service.
//
// EnsureLoadBalancer will not modify service or nodes
func (lb *loadBalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	glog.V(1).Infof("EnsureLoadBalancer: service=%s", service.Name)

	lbStatus, exists, err := lb.GetLoadBalancer(ctx, clusterName, service)
	if err != nil {
		return nil, err
	}

	if !exists {
		lbRequest, err := lb.buildLoadBalancerRequest(service, nodes)
		if err != nil {
			return nil, err
		}

		_, loadBalancer, err := lb.client.CreateLoadBalancer(lbRequest)
		if err != nil {
			return nil, err
		}

		err = lb.client.WaitForState(loadBalancer, lbStatusActive, 30, 60)
		if err != nil {
			return nil, err
		}

		return &v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{
					IP: loadBalancer.Ip,
				},
			},
		}, nil
	}

	err = lb.UpdateLoadBalancer(ctx, clusterName, service, nodes)
	if err != nil {
		return nil, err
	}

	lbStatus, exists, err = lb.GetLoadBalancer(ctx, clusterName, service)
	if err != nil {
		return nil, err
	}

	return lbStatus, nil

}

// UpdateLoadBalancer updates the load balancer for service to balance across
// the servers in nodes.
//
// UpdateLoadBalancer will not modify service or nodes.
func (lb *loadBalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	glog.V(1).Infof("UpdateLoadBalancer: service=%s", service.Name)

	lbName := cloudprovider.GetLoadBalancerName(service)
	loadBalancer, err := lb.lbByName(ctx, lbName)
	if err != nil {
		return err
	}

	serverIpIDs, err := lb.nodesToServerIpIDs(nodes)
	if err != nil {
		return err
	}

	loadBalancer, err = lb.client.AddLoadBalancerServerIps(loadBalancer.Id, serverIpIDs)
	if err != nil {
		return err
	}

	err = lb.client.WaitForState(loadBalancer, lbStatusActive, 30, 60)
	if err != nil {
		return err
	}

	return nil
}

// EnsureLoadBalancerDeleted deletes the specified loadbalancer if it exists.
// nil is returned if the load balancer for service does not exist or is
// successfully deleted.
//
// EnsureLoadBalancerDeleted will not modify service.
func (lb *loadBalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	glog.V(1).Infof("EnsureLoadBalancerDeleted: service=%s", service.Name)
	_, exists, err := lb.GetLoadBalancer(ctx, clusterName, service)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	lbName := cloudprovider.GetLoadBalancerName(service)

	loadBalancer, err := lb.lbByName(ctx, lbName)
	if err != nil {
		return err
	}

	_, err = lb.client.DeleteLoadBalancer(loadBalancer.Id)
	return err
}

// nodesToServerIpIDs returns a []string containing ids of the ips id all servers identified by the
// oneandoneNodeInstanceIdLabel on a node.
//
// oneandoneNodeInstanceIdLabel on nodes are assumed to match oneandone server names.
func (lb *loadBalancer) nodesToServerIpIDs(nodes []*v1.Node) ([]string, error)  {
	var serverIpIDs []string

	for _, node := range nodes {
		server, err := serverFromNode(node, lb.client)
		if err != nil {
			glog.V(1).Infof("nodesToServerIDs: Error looking up servers for matching: %s", node.Name)
			break
		}

		for _, ip := range server.Ips {
			serverIpIDs = append(serverIpIDs, ip.Id)
		}
	}

	return serverIpIDs, nil
}

// buildLoadBalancerRequest returns a *oneandone.LoadBalancerRequest to balance
// requests for service across nodes.
func (lb *loadBalancer) buildLoadBalancerRequest(service *v1.Service, nodes []*v1.Node) (*oneandone.LoadBalancerRequest, error) {
	lbName := cloudprovider.GetLoadBalancerName(service)

	algorithm := getAlgorithm(service)

	forwardingRules, err := buildForwardingRules(service)
	if err != nil {
		return nil, err
	}

	regionId, err := lb.getIdForRegion(lb.region)
	if err != nil {
		return nil, err
	}

	healthCheck := buildHealthCheck(service)

	return &oneandone.LoadBalancerRequest{
		Name:					lbName,
		HealthCheckTest:		healthCheck.CheckTest,
		HealthCheckInterval:	&healthCheck.CheckInterval,
		Persistence:			&healthCheck.Persistence,
		PersistenceTime:		&healthCheck.PersistenceTime,
		DatacenterId:			regionId,
		Method:					algorithm,
		Rules:					forwardingRules,
	}, nil
}

func (lb *loadBalancer) getIdForRegion(regionCode string) (string, error) {
	dcs, err := lb.client.ListDatacenters()
	if err != nil {
		return "", err
	}

	for _, dc := range dcs {
		if dc.CountryCode == regionCode {
			return dc.Id, nil
		}
	}

	return "", fmt.Errorf("getIdForRegion: Cloud not find dataceenter with country_code: %s", regionCode)
}

// buildHealthCheck returns a *HealthCheck helper object which is used for loadbalancer
// creation requests TODO: Get these from service annotations
func buildHealthCheck(service *v1.Service) *HealthCheck {
	return &HealthCheck{
		CheckTest: "ICMP",
		CheckInterval: 15,
		Persistence: true,
		PersistenceTime: 1200,
	}
}

// getAlgorithm returns the load balancing algorithm to use for service.
// ROUND_ROBIN is returned when service does not specify an algorithm.
func getAlgorithm(service *v1.Service) string {
	algo := service.Annotations[oneandoneLoadbalancerMethod]

	switch algo {
	case "LEAST_CONNECTIONS":
		return "LEAST_CONNECTIONS"
	default:
		return "ROUND_ROBIN"
	}
}

// lbByName gets a oneandone Load Balancer by name. The returned error will
// be lbNotFound if the load balancer does not exist.
func (lb *loadBalancer) lbByName(ctx context.Context, name string) (*oneandone.LoadBalancer, error) {
	lbs, err := lb.client.ListLoadBalancers()
	if err != nil {
		return nil, err
	}

	for _, lb := range lbs {
		if lb.Name == name {
			return &lb, nil
		}
	}

	return nil, errLBNotFound
}

// buildForwardingRules returns the forwarding rules of the Load Balancer of
// service.
func buildForwardingRules(service *v1.Service) ([]oneandone.LoadBalancerRule, error) {
	var forwardingRules []oneandone.LoadBalancerRule
	for _, port := range service.Spec.Ports {
		var forwardingRule oneandone.LoadBalancerRule

		forwardingRule.Protocol = string(port.Protocol)

		forwardingRule.PortBalancer = uint16(port.Port)
		forwardingRule.PortServer = uint16(port.NodePort)
		forwardingRule.Source = "0.0.0.0" //TODO: get this from service annotation

		forwardingRules = append(forwardingRules, forwardingRule)
	}

	return forwardingRules, nil
}