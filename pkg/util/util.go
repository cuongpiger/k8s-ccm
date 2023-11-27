package util

import (
	"context"
	"encoding/json"
	"fmt"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	clientset "k8s.io/client-go/kubernetes"
	"time"
)

// MyDuration is the encoding.TextUnmarshaler interface for time.Duration
type MyDuration struct {
	time.Duration
}

// Contains searches if a string list contains the given string or not.
func Contains(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}
	return false
}

// Sprintf255 formats according to a format specifier and returns the resulting string with a maximum length of 255 characters.
func Sprintf255(format string, args ...interface{}) string {
	return CutString255(fmt.Sprintf(format, args...))
}

// CutString255 makes sure the string length doesn't exceed 255, which is usually the maximum string length in OpenStack.
func CutString255(original string) string {
	ret := original
	if len(original) > 255 {
		ret = original[:255]
	}
	return ret
}

// StringListEqual compares two string list, returns true if they have the same items, order doesn't matter
func StringListEqual(list1, list2 []string) bool {
	if len(list1) == 0 && len(list2) == 0 {
		return true
	}

	if len(list1) != len(list2) {
		return false
	}

	s1 := sets.New[string]()
	for _, s := range list1 {
		s1.Insert(s)
	}
	s2 := sets.New[string]()
	for _, s := range list2 {
		s2.Insert(s)
	}

	return s1.Equal(s2)
}

// PatchService makes patch request to the Service object.
func PatchService(ctx context.Context, client clientset.Interface, cur, mod *coreV1.Service) error {
	curJSON, err := json.Marshal(cur)
	if err != nil {
		return fmt.Errorf("failed to serialize current service object: %s", err)
	}

	modJSON, err := json.Marshal(mod)
	if err != nil {
		return fmt.Errorf("failed to serialize modified service object: %s", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJSON, modJSON, coreV1.Service{})
	if err != nil {
		return fmt.Errorf("failed to create 2-way merge patch: %s", err)
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return nil
	}
	_, err = client.CoreV1().Services(cur.Namespace).Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch service object %s/%s: %s", cur.Namespace, cur.Name, err)
	}

	return nil
}
