package v1alpha1

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WordpressSpec defines the desired state of Wordpress
// +k8s:openapi-gen=true
type WordpressSpec struct {
  // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
  // Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
  // Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
  Version string `json:"version"`
}

// WordpressStatus defines the observed state of Wordpress
// +k8s:openapi-gen=true
type WordpressStatus struct {
  // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
  // Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
  // Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Wordpress is the Schema for the wordpresses API
// +k8s:openapi-gen=true
type Wordpress struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`

  Spec   WordpressSpec   `json:"spec,omitempty"`
  Status WordpressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WordpressList contains a list of Wordpress
type WordpressList struct {
  metav1.TypeMeta `json:",inline"`
  metav1.ListMeta `json:"metadata,omitempty"`
  Items           []Wordpress `json:"items"`
}

func init() {
  SchemeBuilder.Register(&Wordpress{}, &WordpressList{})
}
