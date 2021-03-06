package stash

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindRestic = "Restic"
	ResourceNameRestic = "restic"
	ResourceTypeRestic = "restics"

	ResourceKindRecovery = "Recovery"
	ResourceNameRecovery = "recovery"
	ResourceTypeRecovery = "recoveries"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Restic struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResticSpec   `json:"spec,omitempty"`
	Status            ResticStatus `json:"status,omitempty"`
}

type ResticSpec struct {
	Selector      metav1.LabelSelector `json:"selector,omitempty"`
	FileGroups    []FileGroup          `json:"fileGroups,omitempty"`
	Backend       Backend              `json:"backend,omitempty"`
	Schedule      string               `json:"schedule,omitempty"`
	UseAutoPrefix PrefixType           `json:"useAutoPrefix,omitempty"`
	// Pod volumes to mount into the sidecar container's filesystem.
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
	// Compute Resources required by the sidecar container.
	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

type ResticStatus struct {
	FirstBackupTime          *metav1.Time `json:"firstBackupTime,omitempty"`
	LastBackupTime           *metav1.Time `json:"lastBackupTime,omitempty"`
	LastSuccessfulBackupTime *metav1.Time `json:"lastSuccessfulBackupTime,omitempty"`
	LastBackupDuration       string       `json:"lastBackupDuration,omitempty"`
	BackupCount              int64        `json:"backupCount,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ResticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restic `json:"items,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Recovery struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RecoverySpec   `json:"spec,omitempty"`
	Status            RecoveryStatus `json:"status,omitempty"`
}

type RecoverySpec struct {
	Restic     string `json:"restic,omitempty"`
	SnapshotID string `json:"snapshotID,omitempty"`
	// Path       string `json:"path,omitempty"`
	// Host       string `json:"path,omitempty"`
	// target volume where snapshot will be restored
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

type RecoveryStatus struct {
	RecoveryStatus   string `json:"recoveryStatus,omitempty"`
	RecoveryDuration string `json:"recoveryDuration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RecoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Recovery `json:"items,omitempty"`
}

type FileGroup struct {
	// Source of the backup volumeName:path
	Path string `json:"path,omitempty"`
	// Tags of a snapshots
	Tags []string `json:"tags,omitempty"`
	// retention policy of snapshots
	RetentionPolicy RetentionPolicy `json:"retentionPolicy,omitempty"`
}

type Backend struct {
	StorageSecretName string `json:"storageSecretName,omitempty"`

	Local *LocalSpec `json:"local,omitempty"`
	S3    *S3Spec    `json:"s3,omitempty"`
	GCS   *GCSSpec   `json:"gcs,omitempty"`
	Azure *AzureSpec `json:"azure,omitempty"`
	Swift *SwiftSpec `json:"swift,omitempty"`
	// B2    *B2Spec         `json:"b2,omitempty"`
	// Rest  *RestServerSpec `json:"rest,omitempty"`
}

type LocalSpec struct {
	VolumeSource core.VolumeSource `json:"volumeSource,omitempty"`
	Path         string            `json:"path,omitempty"`
}

type S3Spec struct {
	Endpoint string `json:"endpoint,omitempty"`
	Bucket   string `json:"bucket,omiempty"`
	Prefix   string `json:"prefix,omitempty"`
}

type GCSSpec struct {
	Bucket string `json:"bucket,omiempty"`
	Prefix string `json:"prefix,omitempty"`
}

type AzureSpec struct {
	Container string `json:"container,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
}

type SwiftSpec struct {
	Container string `json:"container,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
}

type B2Spec struct {
	Bucket string `json:"bucket,omiempty"`
	Prefix string `json:"prefix,omitempty"`
}

type RestServerSpec struct {
	URL string `json:"url,omiempty"`
}

type RetentionStrategy string

const (
	KeepLast    RetentionStrategy = "--keep-last"
	KeepHourly  RetentionStrategy = "--keep-hourly"
	KeepDaily   RetentionStrategy = "--keep-daily"
	KeepWeekly  RetentionStrategy = "--keep-weekly"
	KeepMonthly RetentionStrategy = "--keep-monthly"
	KeepYearly  RetentionStrategy = "--keep-yearly"
	KeepTag     RetentionStrategy = "--keep-tag"
)

type RetentionPolicy struct {
	KeepLast    int      `json:"keepLast,omitempty"`
	KeepHourly  int      `json:"keepHourly,omitempty"`
	KeepDaily   int      `json:"keepDaily,omitempty"`
	KeepWeekly  int      `json:"keepWeekly,omitempty"`
	KeepMonthly int      `json:"keepMonthly,omitempty"`
	KeepYearly  int      `json:"keepYearly,omitempty"`
	KeepTags    []string `json:"keepTags,omitempty"`
	Prune       bool     `json:"prune,omitempty"`
	DryRun      bool     `json:"dryRun,omitempty"`
}
