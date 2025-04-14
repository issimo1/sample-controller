package v1alpha1

import (
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Lw struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LwSpec   `json:"spec"`
	Status LwStatus `json:"status"`
}

type LwSpec struct {
	TargetRef v1.CrossVersionObjectReference `json:"targetRef"`
	TargetNs  []string                       `json:"targetNs"`
	Replicas  *int32                         `json:"replicas"`
	Data      *map[string]string             `json:"data"`
}

type LwStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LwList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Lw `json:"items"`
}
