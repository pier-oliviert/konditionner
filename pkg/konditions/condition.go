package konditions

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The core type that a Status can interact with. All the public APIs are methods defined on this struct
// and as such, users should interact with Konditionner through this struct, for the most part.
// Conditions should be initialized before being used which can be as easy as defining a non-pointer value for
// the field.
//
//	type Status struct {
//		Conditions konditions.Conditions `json:"conditions"
//	}
//
// At this point you can start using conditions:
//
//	myResource := MyCustomResource{}
//	condition := myResource.Status.Conditions.FindOrInitializeFor(konditions.ConditionType("S3 Bucket")
//	if condition.Status == konditions.ConditionInitialized {
//		err, s3 := s3.Create() // ... Create an s3 bucket ...
//		if err != nil {
//			condition.Status = konditions.ConditionError
//			condition.Reason = err.Error()
//			conditions.SetCondition(condition)
//		}
//
//		condition.Status = konditions.ConditionCreated
//		condition.Reason = fmt.Sprintf("S3 Bucket: %s", s3.Name)
//
//		myResource.Status.Conditions.SetCondition(condition)
//
//	}

type Conditions []Condition

// ConditionType is a user-extendable type that is stored in each Condition
// to define the type of condition the user wants to control. By default, no condition
// type exists and they need to be defined by the user:
//
// type ConditionBucket ConditionType = "S3 Bucket"
// type ConditionVolume ConditionType = "Volume"
// type ConditionPod ConditionType = "Pod Binding"

type ConditionType string

// ConditionStatus is a user-extendable type that is stored in each Condition and add context to
// the ConditionType. Konditionner come with a list of default ConditionStatus that can be used as needed.
// Those ConditionTypes are defined below and more information can be found there as to their usage.

type ConditionStatus string

const (
	// ConditionInitialized is the first status a condition can have, while this value can be set manually
	// by a user, it often is set using the `conditions.FindOrInitializeFor(ConditionType)` helper function. This
	// default value aims at giving user the ability to set a condition themselves with sensible defaults.
	ConditionInitialized ConditionStatus = "Initialized"

	// ConditionCompleted should be used when a condition has ran to completion. This means that no further action
	// are needed for this condition (besides termination, see below).
	ConditionCompleted ConditionStatus = "Completed"

	// ConditionCreated is useful when an external resource needs to be created and configured in multiple
	// reconciliation loop. This state can indicated that some initial work was successfully done but more
	// work needs to be done.
	ConditionCreated ConditionStatus = "Created"

	// ConditionTerminating is useful when a custom resource is marked for deletion but a finalizer
	// was configured on the object. Terminating can be used if the termination requires more than one reconciliation loop.
	// This can be particularly useful when deletion of some external resources is done asynchronously and the user doesn't
	// want to block a reconciliation loop only for assuring that said resource was successfuly created.
	ConditionTerminating ConditionStatus = "Terminating"

	// ConditionTerminated is done to indicate that the finalizer attached to the object is removed and the condition
	// doesn't need to do additional work. A condition that is terminated shouldn't need to be worked on.
	ConditionTerminated ConditionStatus = "Terminated"

	// ConditionError means an error occurred. The error can be of any type. When a condition is marked as errored, it should not
	// be worked on again. Graceful errors should not be marked here but rather be idenfitied through the eventRecorder.
	// Errors are fatal, and the Reason of a condition can be used to store the String() value of the error to help the user
	// know what happened.
	ConditionError ConditionStatus = "Error"

	// ConditionLocked is a special status to tell the reconciler that a condition is being worked on. This is useful when
	// a condition is linked to an external resource (cloud provider, third party, etc.) and you want to have avoid creating
	// duplicate of a resources externally. This can help avoid multiple reconciliation happen at a same time. A reconciliation
	// should attempt to `Locked` the resource first, if the update/patch is succesful, then it means this reconciliation loop
	// as acquired a lock on this condition. It is important to note, however, that it's not a "real" lock. We're in a distributed system and
	// the etcd/kubernetes client interaction include layers of caching and logic.
	ConditionLocked ConditionStatus = "Locked"
)

// Condition is an individual condition that makes the Conditions type. Each of those conditions are created
// to isolate some behavior the user wants control over.
type Condition struct {
	// The type of the condition you want to have control over. The type is a user-defined value that extends the ConditionType. The type
	// serves as a way to identify the condition and it can be fetched from the Conditions type by using any of the finder methods.
	// ---
	// +required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=316
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type"`

	// Current status of the condition. This field should mutate over the lifetime of the condition. By default, it starts as
	// ConditionInitialized and it's up to the user to modify the status to reflect where the condition is, relative to its lifetime.
	// ---
	// +required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=128
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`

	// LastTransitionTime is the last time the condition transitioned from one status to another. This value is set automatically by
	// the Conditions' method and as such, don't need to be set by the user.
	// ---
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime meta.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`

	// Reason represents the details about the transition and its current state.
	// For instance, it can hold the description of an error.Error() if the status is set to
	// ConditionError. This field is optional and should be used to give additionnal context.
	// Since this value can be overriden by future changes to the status of the condition,
	// users might want to also record the Reason through Kubernete's EventRecorder.
	// ---
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
}

// Helper function that returns true if the Status of the condition is equal
// to one of the statuses provided.
func (c Condition) StatusIsOneOf(statuses ...ConditionStatus) bool {
	for _, s := range statuses {
		if c.Status == s {
			return true
		}
	}

	return false
}

// Kubernetes requires any struct that can be stored in a Custom Resource Definition(CRD) to
// implement these DeepCopy functions. They aren't interfaces as the arguments and return values
// are explicitly typed. Usually, when using tools like kube-builder/controller-runtime, those functions
// are auto-generated. Because Konditionner is not a CRD, those functions needs to exists here.

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conditions) DeepCopyInto(out *Conditions) {
	for i := range *in {
		(*in)[i].DeepCopyInto(&(*out)[i])
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Status.
func (in *Conditions) DeepCopy() Conditions {
	if in == nil {
		return nil
	}
	out := make(Conditions, len(*in))
	in.DeepCopyInto(&out)
	return out
}

func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}
