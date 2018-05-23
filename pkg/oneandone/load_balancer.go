package oneandone

import (
	"context"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
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

	// defaultActiveCheckTick is the number of seconds between load balancer
	// status checks when waiting for activation.
	defaultActiveCheckTick = 5

	// statuses for oneandone load balancer
	lbStatusNew     = "new"
	lbStatusActive  = "ACTIVE"
	lbStatusErrored = "errored"
)

type loadBalancer struct {
	client            *oneandone.API
	region            string
	lbActiveTimeout   int
	lbActiveCheckTick int
}

// newLoadbalancers returns a cloudprovider.LoadBalancer whose concrete type is a *loadbalancer.
func newLoadbalancer(client *oneandone.API, region string) cloudprovider.LoadBalancer {
	return &loadBalancer{client, region, defaultActiveTimeout, defaultActiveCheckTick}
}

func (lb *loadBalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	glog.V(1).Infof("GetLoadBalancer: service=%s", service.Name)
	return nil, false, nil
}

func (lb *loadBalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	glog.V(1).Infof("EnsureLoadBalancer: service=%s", service.Name)
	return &v1.LoadBalancerStatus{}, nil
}

func (lb *loadBalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	glog.V(1).Infof("UpdateLoadBalancer: service=%s", service.Name)
	return nil
}

func (lb *loadBalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	glog.V(1).Infof("EnsureLoadBalancerDeleted: service=%s", service.Name)
	return nil
}
