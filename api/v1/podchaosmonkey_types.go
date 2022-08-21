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
	// If not specified, all fields in the `PodChaosMonkeyConditions` are set to their
	// default values.
	// +optional
	Conditions *PodChaosMonkeyConditions `json:"conditions,omitempty" protobuf:"bytes,2,opt,name=conditions"`
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// Kind of the target; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string `json:"kind" protobuf:"bytes,1,name=kind"`
	// Name of the target; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	// API version of the target
	APIVersion string `json:"apiVersion" protobuf:"bytes,3,name=apiVersion"`
}

// PodChaosMonkeyPolicy describes the rules on how to apply chaos pod eviction.
type PodChaosMonkeyConditions struct {
	// Minimal number of pods which need to be alive
	// +optional
	MinPods *int32 `json:"minPods,omitempty" protobuf:"varint,1,opt,name=minPods"`
	// Minimal time in minutes that has to be running
	MinTimeRunning *int32 `json:"minTimeRunning" protobuf:"varint,2,name=minRunning"`
	// Minimal Random time in minutes between two consecutive pod evictions
	MinTimeRandom *int32 `json:"minTimeRandom" protobuf:"varint,3,name=minTimeRandom"`
	// Maximum Random time in minutes between two consecutive pod evictions
	MaxTimeRandom *int32 `json:"maxTimeRandom" protobuf:"varint,4,name=maxTimeRandom"`
	// Grace period for pod termination
	// +optional
	GracePeriodSeconds *int64 `json:"gracePeriodSeconds,omitempty" protobuf:"varint,5,opt,name=gracePeriodSeconds"`
}

// PodChaosMonkeyStatus defines the observed state of PodChaosMonkey
type PodChaosMonkeyStatus struct {
	// Last time the chaos monkey evicted a pod
	LastEvictionAt *metav1.Time `json:"lastEvictionAt,omitempty" protobuf:"bytes,1,opt,name=lastEvictionAt"`
	// Number of times the chaos monkey has been executed
	NumberOfEvictions *int32 `json:"numberOfEvictions,omitempty" protobuf:"varint,2,opt,name=numberOfEvictions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=chaos
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
