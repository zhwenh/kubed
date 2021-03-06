package util

import (
	"strconv"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func IsPreferredAPIResource(kubeClient kubernetes.Interface, groupVersion, kind string) bool {
	if resourceList, err := kubeClient.Discovery().ServerPreferredResources(); err == nil {
		for _, resources := range resourceList {
			if resources.GroupVersion != groupVersion {
				continue
			}
			for _, resource := range resources.APIResources {
				if resources.GroupVersion == groupVersion && resource.Kind == kind {
					return true
				}
			}
		}
	}
	return false
}

const (
	MaxSyncInterval = 5 * time.Minute
)

func IsRecent(t metav1.Time) bool {
	return time.Now().Sub(t.Time) < MaxSyncInterval
}

func ObfuscateSecret(in core.Secret) *core.Secret {
	data := make(map[string][]byte)
	for k := range in.Data {
		data[k] = []byte("-")
	}
	in.Data = data
	return &in
}

func GetBool(m map[string]string, key string) (bool, error) {
	if m == nil {
		return false, nil
	}
	v, ok := m[key]
	if !ok || v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}

func GetString(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}
