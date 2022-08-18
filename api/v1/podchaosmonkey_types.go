/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodChaosMonkeySpec defines the desired state of PodChaosMonkey
type PodChaosMonkeySpec struct {
	// TargetRef points to the controller managing the set of pods for the testing
	TargetRef *CrossVersionObjectReference `json:"targetRef" protobuf:"bytes,1,name=targetRef"`

	// Describes the rules on how chaos testing are applied to the pods.
	// If not specified, all fields in the `PodChaosMonkeyPolicy` are set to their
	// default values.
	// +optional
	Conditions *PodChaosMonkeyConditions `json:"conditions,omitempty" protobuf:"bytes,2,opt,name=conditions"`
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string
	// API version of the referent
	// +optional
	APIVersion string
}

// PodChaosMonkeyPolicy describes the rules on how to apply chaos pod eviction.
type PodChaosMonkeyConditions struct {
	// Minimal time in hours that has to be running
	// +optional
	MinRunning *float64 `json:"minRunning,omitempty" protobuf:"varint,1,opt,name=minRunning"`
	// Minimal number of pods which need to be alive
	// +optional
	MinPods *int32 `json:"minPods,omitempty" protobuf:"varint,2,opt,name=minPods"`
}

// PodChaosMonkeyStatus defines the observed state of PodChaosMonkey
type PodChaosMonkeyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PodChaosMonkey is the Schema for the podchaosmonkeys API
type PodChaosMonkey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodChaosMonkeySpec   `json:"spec,omitempty"`
	Status PodChaosMonkeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PodChaosMonkeyList contains a list of PodChaosMonkey
type PodChaosMonkeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodChaosMonkey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodChaosMonkey{}, &PodChaosMonkeyList{})
}
