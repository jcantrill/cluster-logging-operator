/*
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

package v1beta1

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLogDistributorSpec defines the desired state of ClusterLogDistributor
type ClusterLogDistributorSpec struct {

	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator.
	//
	// +kubebuilder:default:=Managed
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Management State"
	ManagementState obsv1.ManagementState `json:"managementState,omitempty"`

	// Specification of the Collector deployment to define
	// resource limits and workload placement
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Resources and Placement",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:advanced"}
	Collector *obsv1.CollectorSpec `json:"collector,omitempty"`

	// Priority is the priority of this ClusterLogDistributor in relation to others.
	// A larger number will take precedence when evaluating which ClusterLogDistribor will service
	// a LogForwarder
	Priority int `json:"priority,omitempty"`

	// CollectionPolicy is the policy that defines which logs are to be collected.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CollectionPolicy"
	CollectionPolicy ClusterLogDistributorCollectionPolicySpec `json:"collectionPolicy,omitempty"`
}

type ClusterLogDistributorCollectionPolicySpec struct {

	// Container is the collection policy for container sources
	// +kubebuilder:validation:Required
	Container *ClusterLogDistributorContainerInputSpec `json:"container,omitempty"`
}

type ClusterLogDistributorContainerInputSpec struct {

	// Includes is the set of namespaces and containers to include when collecting logs.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Include"
	Includes []obsv1.NamespaceContainerSpec `json:"includes,omitempty"`

	// Excludes is the set of namespaces and containers to ignore when collecting logs.
	//
	// Takes precedence over Includes option.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Exclude"
	Excludes []obsv1.NamespaceContainerSpec `json:"excludes,omitempty"`

	// NamespaceMatchLabels for logs from namespaces with matching labels.
	//
	// Only logs from pods from namespaces that match this label selector are collected.  This selector
	// is further restricted to included or excluded namespaces.
	//
	// If absent or empty, logs are collected regardless of labels.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pod NamespaceMatchLabels",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Pod"}
	NamespaceMatchLabels map[string]string `json:"namespaceMatchLabels,omitempty"`
}

// ClusterLogDistributorStatus defines the observed state of ClusterLogDistributor
type ClusterLogDistributorStatus struct {
	// Conditions of the log forwarder.
	//e=status
	//	// +operator-sdk:csv:customresourcedefinitions:typ,displayName="ClusterLogDistributor Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ClusterLogDistributor is an API to configure forwarding logs.
//
// You configure forwarding by specifying a list of `pipelines`,
// which forward from a set of named inputs to a set of named outputs.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=observability,shortName=obscld;cld,scope=Cluster
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z][a-z0-9-]{1,61}[a-z0-9]$')",message="Name must be a valid DNS1035 label"
type ClusterLogDistributor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLogDistributorSpec   `json:"spec,omitempty"`
	Status ClusterLogDistributorStatus `json:"status,omitempty"`
}

// ClusterLogDistributorList contains a list of ClusterLogDistributor
//
// +kubebuilder:object:root=true
type ClusterLogDistributorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogDistributor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogDistributor{}, &ClusterLogDistributorList{})
}
