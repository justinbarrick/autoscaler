/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package digitalocean

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

const (
	// ProviderNameDO is the name of DO cloud provider.
	ProviderNameDO = "do"
)

const (
	maxAutoprovisionedSize = 1000
	minAutoprovisionedSize = 0
)

// Big machines are temporarily commented out.
// TODO(mwielgus): get this list programmatically
var autoprovisionedMachineTypes = []string{
	"n1-standard-1",
	"n1-standard-2",
	"n1-standard-4",
	"n1-standard-8",
	"n1-standard-16",
	//"n1-standard-32",
	//"n1-standard-64",
	"n1-highcpu-2",
	"n1-highcpu-4",
	"n1-highcpu-8",
	"n1-highcpu-16",
	//"n1-highcpu-32",
	// "n1-highcpu-64",
	"n1-highmem-2",
	"n1-highmem-4",
	"n1-highmem-8",
	"n1-highmem-16",
	//"n1-highmem-32",
	//"n1-highmem-64",
}

// DOCloudProvider implements CloudProvider interface.
type DOCloudProvider struct {
	doManager DOManager

	// This resource limiter is used if resource limits are not defined through cloud API.
	resourceLimiterFromFlags *cloudprovider.ResourceLimiter
}

// BuildGkeCloudProvider builds CloudProvider implementation for GKE.
func BuildDOCloudProvider(doManager DOManager, resourceLimiter *cloudprovider.ResourceLimiter) (*DOCloudProvider, error) {
	return &DOCloudProvider{doManager: doManager, resourceLimiterFromFlags: resourceLimiter}, nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (do *DOCloudProvider) Cleanup() error {
	return cloudprovider.ErrNotImplemented
}

// Name returns name of the cloud provider.
func (do *DOCloudProvider) Name() string {
	return ProviderNameDO
}

// NodeGroups returns all node groups configured for this cloud provider.
func (do *DOCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return do.doManager.GetNodeGroups()
}

// NodeGroupForNode returns the node group for the given node.
func (do *DOCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (do *DOCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (do *DOCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return autoprovisionedMachineTypes, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (do *DOCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	nodePoolName := fmt.Sprintf("pool-%s-%d", machineType, time.Now().Unix())

	pool := &doPool{
		doManager:       do.doManager,
		autoprovisioned: true,
		exist:           false,
		nodePoolName:    nodePoolName,
		minSize:         minAutoprovisionedSize,
		maxSize:         maxAutoprovisionedSize,
	}

	return pool, nil
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (do *DOCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (do *DOCloudProvider) Refresh() error {
	return cloudprovider.ErrNotImplemented
}

// GetClusterInfo returns the project id, location and cluster name.
func (do *DOCloudProvider) GetClusterInfo() (projectId, location, clusterName string) {
	return "default", "tor1", ""
}

type doPool struct {
	doManager DOManager

	id string
	minSize         int
	maxSize         int
	autoprovisioned bool
	exist           bool
	nodePoolName    string
}

// MaxSize returns maximum size of the node group.
func (do *doPool) MaxSize() int {
	return do.maxSize
}

// MinSize returns minimum size of the node group.
func (do *doPool) MinSize() int {
	return do.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (do *doPool) TargetSize() (int, error) {
	return 0, cloudprovider.ErrNotImplemented
}

// IncreaseSize increases node pool size
func (do *doPool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
/*
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size)+delta > mig.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, mig.MaxSize())
	}
	return mig.gkeManager.SetMigSize(mig, size+int64(delta))
*/
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (do *doPool) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
/*
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	nodes, err := mig.gkeManager.GetMigNodes(mig)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return mig.gkeManager.SetMigSize(mig, size+int64(delta))
*/
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the nodes from the group.
func (mig *doPool) DeleteNodes(nodes []*apiv1.Node) error {
/*
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size) <= mig.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*gce.GceRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := mig.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different mig than %s", node.Name, mig.Id())
		}
		gceref, err := gce.GceRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, gceref)
	}
	return mig.gkeManager.DeleteInstances(refs)
*/
	return cloudprovider.ErrNotImplemented
}

// Id returns node pool id.
func (do *doPool) Id() string {
	return do.id
}

// Debug returns a debug string for the node pool.
func (do *doPool) Debug() string {
	return ""
}

// Nodes returns a list of all nodes that belong to this node group.
func (do *doPool) Nodes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (do *doPool) Exist() bool {
	return false
}

// Create creates the node group on the cloud provider side.
func (do *doPool) Create() (cloudprovider.NodeGroup, error) {
	if !do.exist && do.autoprovisioned {
	}
	return nil, fmt.Errorf("Cannot create non-autoprovisioned node group")
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (do *doPool) Delete() error {
	if do.exist && do.autoprovisioned {
	}
	return fmt.Errorf("Cannot delete non-autoprovisioned node group")
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (do *doPool) Autoprovisioned() bool {
	return do.autoprovisioned
}

// TemplateNodeInfo returns a node template for this node group.
func (do *doPool) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
/*
	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(mig.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
*/
	return nil, cloudprovider.ErrNotImplemented
}

var _ cloudprovider.NodeGroup = (*doPool)(nil)
