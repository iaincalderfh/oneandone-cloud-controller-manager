package oneandone

import (
	"io"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
)

const (
	ProviderName = "oneandone"
)

func init() {
	cloudprovider.RegisterCloudProvider("oneandone", func(config io.Reader) (cloudprovider.Interface, error) {
		return &CloudProvider{lb: &loadBalancer{}}, nil
	})
}

// Compile-time check that CloudProvider implements cloudprovider.Interface
var _ cloudprovider.Interface = &CloudProvider{}

type CloudProvider struct {
	lb cloudprovider.LoadBalancer
}

func (cp *CloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {
	glog.V(1).Infof("%s CloudProvider: Intialize called", ProviderName)
}

func (cp *CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	glog.V(1).Infof("%s CloudProvider: LoadBalancer called", ProviderName)
	return cp.lb, true
}

func (cp *CloudProvider) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

func (cp *CloudProvider) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (cp *CloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (cp *CloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (cp *CloudProvider) ProviderName() string {
	return ProviderName
}

func (cp *CloudProvider) HasClusterID() bool {
	return false
}
