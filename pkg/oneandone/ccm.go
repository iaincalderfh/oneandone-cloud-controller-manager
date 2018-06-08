package oneandone

import (
	"io"

	"fmt"
	"os"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/leroyshirtoFH/oneandone-cloudserver-sdk-go"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
)

const (
	oneAPITokenEnv       = "ONEANDONE_API_KEY"
	oneOverrideAPIURLEnv = "ONEANDONE_OVERRIDE_URL"
	oneInstanceRegionEnv = "ONEANDONE_INSTANCE_REGION"
	providerName         = "oneandone"
)

func init() {
	cloudprovider.RegisterCloudProvider("oneandone", func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := readConfig(config)
		if err != nil {
			return nil, err
		}
		return newCloudProvider(cfg)
	})
}

type cloudProvider struct {
	client       *oneandone.API
	kubeclient   clientset.Interface
	loadbalancer cloudprovider.LoadBalancer
	instances    *instances
}

func newCloudProvider(config *config) (cloudprovider.Interface, error) {
	apiClient, err := newOneAndOneAPIClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to create oneandone API client: %s", err)
	}

	region := os.Getenv(oneInstanceRegionEnv)
	if region == "" {
		return nil, fmt.Errorf("Environment variable %q is required", oneInstanceRegionEnv)
	}

	return &cloudProvider{
		client:       apiClient,
		loadbalancer: newLoadbalancer(apiClient, region),
		instances:    newInstances(apiClient, region),
	}, nil
}

func newOneAndOneAPIClient() (*oneandone.API, error) {
	token := oneandone.SetToken(os.Getenv(oneAPITokenEnv))
	if token == "" {
		return nil, fmt.Errorf("Environment variable %q is required", oneAPITokenEnv)
	}

	baseURL := oneandone.BaseUrl
	if overrideURL := os.Getenv(oneOverrideAPIURLEnv); overrideURL != "" {
		baseURL = overrideURL
	}

	apiClient := oneandone.New(token, baseURL)
	if apiClient == nil {
		return nil, fmt.Errorf("Failed to create oneandone api client")
	}

	return apiClient, nil
}

// Initialize implements cloudprovider.Interface.Initialize
// A Kubernetes client is obtained from the clientBuilder.
func (cp *cloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {
	var err error
	cp.kubeclient, err = clientBuilder.Client("cloud-controller-manager")
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to create kubeclient: %v", err))
	}

	cp.instances.kubeNodesAPI = cp.kubeclient.CoreV1().Nodes()
}

// LoadBalancer implements cloudprovider.Interface.LoadBalancer
// Returns a cloudprovider.LoadBalancer interface for 1AND1.
func (cp *cloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return cp.loadbalancer, true
}

// Instances implements cloudprovider.Interface.Instances
// Returns a cloudprovider.Instances interface for 1AND1.
func (cp *cloudProvider) Instances() (cloudprovider.Instances, bool) {
	return cp.instances, true
}

// Zones implements cloudprovider.Interface.Zones
// Returns (nil,false) as this interface is not currently supported for 1AND1.
func (cp *cloudProvider) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

// Clusters implements cloudprovider.Interface.Clusters
// Returns (nil,false) as this interface is not currently supported for 1AND1.
func (cp *cloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes implements cloudprovider.Interface.Routes
// Returns (nil,false) as this interface is not currently supported for 1AND1.
func (cp *cloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ProviderName implements cloudprovider.Interface.ProviderName
// Returns our cloud provider name.
func (cp *cloudProvider) ProviderName() string {
	return providerName
}

// HasClusterID implements cloudprovider.Interface.HasClusterID
// Returns false.
func (cp *cloudProvider) HasClusterID() bool {
	return false
}
