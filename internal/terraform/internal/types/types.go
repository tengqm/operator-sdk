// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type TFConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TFConfig `json:"items"`
}

type TFConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TFConfigSpec   `json:"spec"`
	Status            TFConfigStatus `json:"status,omitempty"`
}

type TFConfigSpec map[string]interface{}

type TFConfigConditionType string
type ConditionStatus string
type TFConfigConditionReason string

type TFConfigCondition struct {
	Type    TFConfigConditionType   `json:"type"`
	Status  ConditionStatus         `json:"status"`
	Reason  TFConfigConditionReason `json:"reason,omitempty"`
	Message string                  `json:"message,omitempty"`

	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type TFDeploy struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

const (
	ConditionInitialized    TFConfigConditionType = "Initialized"
	ConditionDeployed       TFConfigConditionType = "Deployed"
	ConditionConfigFailed   TFConfigConditionType = "ConfigFailed"
	ConditionIrreconcilable TFConfigConditionType = "Irreconcilable"

	StatusTrue    ConditionStatus = "True"
	StatusFalse   ConditionStatus = "False"
	StatusUnknown ConditionStatus = "Unknown"

	ReasonCreateSuccessful TFConfigConditionReason = "CreateSuccessful"
	ReasonUpdateSuccessful TFConfigConditionReason = "UpdateSuccessful"
	ReasonDeleteSuccessful TFConfigConditionReason = "DeleteSuccessful"
	ReasonCreateError      TFConfigConditionReason = "CreateError"
	ReasonUpdateError      TFConfigConditionReason = "UpdateError"
	ReasonReconcileError   TFConfigConditionReason = "ReconcileError"
	ReasonDeleteError      TFConfigConditionReason = "DeleteError"
)

type TFConfigStatus struct {
	Conditions     []TFConfigCondition `json:"conditions"`
	DeployedConfig *TFDeploy           `json:"deployedConfig,omitempty"`
}

func (s *TFConfigStatus) ToMap() (map[string]interface{}, error) {
	var out map[string]interface{}
	jsonObj, err := json.Marshal(&s)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonObj, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetCondition sets a condition on the status object. If the condition already
// exists, it will be replaced. SetCondition does not update the resource in
// the cluster.
func (s *TFConfigStatus) SetCondition(condition TFConfigCondition) *TFConfigStatus {
	now := metav1.Now()
	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			if s.Conditions[i].Status != condition.Status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = s.Conditions[i].LastTransitionTime
			}
			s.Conditions[i] = condition
			return s
		}
	}

	// If the condition does not exist, initialize the lastTransitionTime
	condition.LastTransitionTime = now
	s.Conditions = append(s.Conditions, condition)
	return s
}

// RemoveCondition removes the condition with the passed condition type from
// the status object. If the condition is not already present, the returned
// status object is returned unchanged. RemoveCondition does not update the
// resource in the cluster.
func (s *TFConfigStatus) RemoveCondition(conditionType TFConfigConditionType) *TFConfigStatus {
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			s.Conditions = append(s.Conditions[:i], s.Conditions[i+1:]...)
			return s
		}
	}
	return s
}

// StatusFor safely returns a typed status block from a custom resource.
func StatusFor(cr *unstructured.Unstructured) *TFConfigStatus {
	switch s := cr.Object["status"].(type) {
	case *TFConfigStatus:
		return s
	case map[string]interface{}:
		var status *TFConfigStatus
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(s, &status); err != nil {
			return &TFConfigStatus{}
		}
		return status
	default:
		return &TFConfigStatus{}
	}
}
