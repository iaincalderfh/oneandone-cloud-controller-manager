package oneandone

import (
	"context"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

// Compile-time check that loadBalancer implements cloudprovider.LoadBalancer
var _ cloudprovider.LoadBalancer = &loadBalancer{}

type loadBalancer struct {
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
