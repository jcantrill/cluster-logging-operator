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

// LogForwarderSpec defines the desired state of LogForwarder
type LogForwarderSpec struct {

	// Specification of the Forwarder deployment to define
	// resource limits and workload placement
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Resources and Placement",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:advanced"}
	Forwarder *obsv1.CollectorSpec `json:"forwarder,omitempty"`

	// Input is the spec of logs to forward
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Input"
	Input LogForwarderInputSpec `json:"input,omitempty"`

	// Outputs are named destinations for log messages.
	//
	// +kubebuilder:validation:Required
	// +listType:=map
	// +listMapKey:=name
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Outputs"
	Outputs []obsv1.OutputSpec `json:"outputs,omitempty"`

	// ServiceAccount points to the ServiceAccount resource used by the forwarder pods and defaults
	// to the service account created for the namespace if not otherwise defined.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account"
	ServiceAccount *obsv1.ServiceAccount `json:"serviceAccount,omitempty"`
}

// Application workload log selector.
// All conditions in the selector must be satisfied (logical AND) to select logs.
type LogForwarderInputSpec struct {
	// The spec for collecting container logs
	//
	// +nullable
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Logs Input"
	Container *ContainerInputSpec `json:"container,omitempty"`

	// CollectionPolicy to apply when collecting logs.
	//
	// Absence of a spec will utilize no preferences when evaluating the LogForwarder
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Collection Policy"
	CollectionPolicy *CollectionPolicy `json:"collectionPolicy,omitempty"`
}

// CollectionPolicy provides hits for the strategy to use when logs are collected.  This can be used to identify
// a prioritized collection agent to address fair share or log distribution as related to other workloads on the
// cluster which utilize a LogForwarder
type CollectionPolicy struct {

	// DistributionClass is the service class to use when evaluating the collection policy
	DistributionClass string `json:"distributionClass,omitempty"`
}

type ContainerInputSpec struct {
	// Selector for logs from pods with matching labels.
	//
	// Only messages from pods with these labels are collected.
	//
	// If absent or empty, logs are collected regardless of labels.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pod Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Pod"}
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Includes is the set of containers to include when collecting logs.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Includes"
	Includes []string `json:"includes,omitempty"`

	// Excludes is the set of containers to ignore when collecting logs.
	//
	// Takes precedence over Includes option.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Excludes"
	Excludes []string `json:"excludes,omitempty"`
}

// LogForwarderStatus defines the observed state of LogForwarder
type LogForwarderStatus struct {
	// Conditions of the log forwarder.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Forwarder Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// LogForwarder is an API to configure forwarding logs.
//
// You configure forwarding by specifying a list of `pipelines`,
// which forward from a set of named inputs to a set of named outputs.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=observability,shortName=obslf;lf
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z][a-z0-9-]{1,61}[a-z0-9]$')",message="Name must be a valid DNS1035 label"
type LogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogForwarderSpec   `json:"spec,omitempty"`
	Status LogForwarderStatus `json:"status,omitempty"`
}

// LogForwarderList contains a list of LogForwarder
//
// +kubebuilder:object:root=true
type LogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogForwarder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogForwarder{}, &LogForwarderList{})
}
