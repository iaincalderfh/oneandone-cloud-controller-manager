package oneandone

import (
	"io"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"os"
	"fmt"
)

const (
	oneAPITokenEnv  = "ONE_API_TOKEN"
	oneOverrideAPIURLEnv = "ONE_OVERRIDE_URL"
	oneInstanceReqionEnv = "ONE_INSTANCE_REGION"
	ProviderName = "oneandone"
)

func init() {
	cloudprovider.RegisterCloudProvider("oneandone", func(config io.Reader) (cloudprovider.Interface, error) {
		return newCloudProvider(config)
	})
}

// Compile-time check that CloudProvider implements cloudprovider.Interface
var _ cloudprovider.Interface = &CloudProvider{}

type CloudProvider struct {
	client        *oneandone.API
	loadbalancer cloudprovider.LoadBalancer
}

func newCloudProvider(config io.Reader) (cloudprovider.Interface, error) {
	token := oneandone.SetToken(os.Getenv(oneAPITokenEnv))
	if token == "" {
		return nil, fmt.Errorf("environment variable %q is required", oneAPITokenEnv)
	}

	baseUrl := oneandone.BaseUrl
	if overrideURL := os.Getenv(oneOverrideAPIURLEnv); overrideURL != "" {
		baseUrl = overrideURL
	}

	apiClient := oneandone.New(token, baseUrl)
	if apiClient == nil {
		return nil, fmt.Errorf("failed to create oneandone api client")
	}

	region := os.Getenv(oneInstanceReqionEnv)
	if region == "" {
		return nil, fmt.Errorf("environment variable %q is required", oneInstanceReqionEnv)
	}

	return &CloudProvider{
		client: apiClient,
		loadbalancer: newLoadbalancer(apiClient, region),
	}, nil
}

func (cp *CloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {
	glog.V(1).Infof("%s CloudProvider: Intialize called", ProviderName)
}

func (cp *CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	glog.V(1).Infof("%s CloudProvider: LoadBalancer called", ProviderName)
	return cp.loadbalancer, true
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
