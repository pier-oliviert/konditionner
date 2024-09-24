package konditions

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var LockNotReleasedErr = errors.New("Condition still locked after task executed. The task needs to set the condition status before returning")

// Lock is and advisory lock that can be used to make sure you have control over a condition
// before running a task that would create external resources. Even though
// this is named a Lock, be aware that we're working in a distributed system
// and the lock is at the application level. This does not have strong atomic guarantees
// but it doesn't mean it's not useful either.
//
// Kubernetes operate cache layers where client of the Kubernetes API can sometime
// operate on a stale cache. This would usually result in an error when *updating* the custom
// resource (or its status):
//
//	var res Resource
//	if err := reconciler.Get(ctx, &res, request); err != nil {
//		// No error on fetching, hit the cache, but the cache is stale
//	}
//
//	bucket, err := createBucketForResource(ctx, res)
//	if err != nil {
//		// No error here
//	}
//
//	res.Status.BucketName = bucket.Name
//	res.Status.Conditions.SetCondition(Condition{
//		Type: ConditionType("Bucket"),
//		Status: konditions.ConditionCreated,
//		Reason: "Bucket Created",
//	})
//
//	if err := reconciler.Status().Update(ctx, &res); err != nil {
//		// The cache was stale, conflict on update..
//		// !Boom!
//	}
//
// In the example above, the error happens after all the work was executed, which leaves
// the user with fetching the resource again, and trying to eventually get a fresh copy of the
// resource so you can make the update, this means you have wrap most of the conciler logic in a retry
// block, making the whole reconciliation process harder to reason about. It also have the downside of
// blocking the execution of the reconciliation loop, which can cause congestion in your pipeline.
//
// This is where the Lock comes handy. Before any work starts, it will attempt to modify
// the condition's Status to `ConditionLocked`. If the condition is successfully updated,
// it means the cache is not stale at the time of updating the condition and it should be safe
// to start working on the Task.
//
// It is the user's job to set the condition at the end of the Task to the end state desired, but
// the Lock will operate on any error that the Task returns: You don't have to update the condition
// on errors.
//
//	var res Resource
//	if err := reconciler.Get(ctx, &res, request); err != nil {
//		// No error on fetching, hit the cache, but the cache is stale
//	}
//
//	lock := konditions.NewLock(res, reconciler.Client, ConditionType("Bucket"))
//	lock.Execute(ctx, func(condition Condition) error {
//		bucket, err := createBucketForResource(ctx, &res)
//		if err != nil {
//			return err
//		}
//
//		res.Status.BucketName = bucket.Name
//	  condition.Status = konditions.ConditionCreated
//		condition.Reason = "Bucket Created"
//		res.Status.Conditions.SetCondition(condition)
//
//		return reconciler.Status().Update(ctx, &res)
//	})
//
// The condition with type "Bucket" will have its Status go through a few stages:
//   - Initialized
//   - Locked
//   - Created *or* Error
type Lock struct {
	client    client.Client
	obj       ConditionalResource
	condition Condition
}

// Task is a unit of work on a given Condition as specified by the lock.
// The condition is a copy of the condition *before* the lock was obtained. This is useful
// as the status can be useful to make a decision.
//
//	lock := konditions.NewLock(res, ConditionType("Bucket"))
//	lock.Execute(ctx, reconciler.Client, func(condition Condition) error {
//		// Even if the condition status is currently `ConditionLocked`, the condition passed
//		// will be a copy with the Status prior to locking it.
//		// In this example, the condition didn't exist, and as such, `FindOrInitializeFor`
//		// returned the condition with the default value: ConditionInitialized
//
//		if condition.Status == ConditionTerminating {
//			if err := deleteBucketForResource(&res); err != nil {
//				return err
//			}
//			condition.Status = ConditionTerminated
//			condition.Reason = "Bucket deleted"
//			res.Status.Conditions.SetCondition(condition)
//
//			return reconciler.Status().Update(ctx, &res)
//		}
//
//		if condition.Status == ConditionInitialized {
//			bucket, err := createBucketForResource(ctx, &res)
//			if err != nil {
//				return err
//			}
//
//			res.Status.BucketName = bucket.Name
//			condition.Status = konditions.ConditionCreated
//			condition.Reason = "Bucket Created"
//			res.Status.Conditions.SetCondition(condition)
//
//			return reconciler.Status().Update(ctx, &res)
//		}
//	})
type Task func(Condition) error

// ConditionObject is an interface for your CRD with the added method Conditions() defined by the
// user. This interface exists to simplify the usage of Lock and can be implemented by adding the conditions getter
// your CRD.
//
//	type MyCRDSpec struct { ... }
//	type MyCRDStatus struct {
//		// ... Other fields ...
//
//		Conditions konditions.Conditions `json:"conditions"`
//	}
//
//	type MyCRD struct {
//		metav1.TypeMeta   `json:",inline"`
//		metav1.ObjectMeta `json:"metadata,omitempty"`
//
//		Spec   DNSRecordSpec   `json:"spec,omitempty"`
//		Status DNSRecordStatus `json:"status,omitempty"`
//	}
//
//	func (m MyCRD) Conditions() *konditions.Conditions {
//		return &m.Status.Conditions
//	}
type ConditionalResource interface {
	Conditions() *Conditions

	client.Object
}

// NewLock returns a fully configured lock to run.
// Providing the CRD as a ConditionalResource and the condition type
// you want to operate on, the lock will fetch the condition, and configure
// itself to be executed when needed.
//
// The lock will hold a copy of the condition with ConditionType at the time
// of its initialized.
//
// The Client interface is usually the reconciler controller you are within.
//
//	lock := konditions.NewLock(res, reconciler.Client, ConditionType("Bucket"))
func NewLock(obj ConditionalResource, c client.Client, ct ConditionType) *Lock {
	condition := obj.Conditions().FindOrInitializeFor(ct)

	return &Lock{
		client:    c,
		condition: condition,
		obj:       obj,
	}
}

// Execute the task after successfully setting the condition to ConditionLocked.
// Calling Execute will attempt to change the condition's Status to ConditionLocked.
// If successful, it will then call Task(condition) where the condition is a copy of
// the condition when the Lock was initialized, this means that even if the condition is "locked",
// the condition passed will have the status *before* it was locked, giving the opportunity to
// the task to analyze what the sItus of the condition was.
//
// If the task returns an error, the condition will be updated to ConditionError and the Reason
// will be set to the error.Error().
//
// It is up to the Task to set the condition to its final state with the appropriate reason.
// Once the task has returned, the Lock will update the status' subresource of the custom resource.
//
// If any error happens while communicating with the Kubernetes API, it will be returned.
// If it were to happen, the condition will not be updated, the error can then be passed to
// the reconciler so it retries the reconciliation loop.
//
//	if err := lock.Execute(ctx, task); err != nil {
//		return ctrl.Result{}, err
//	}
//
// The Execution loop will always return Kubernetes API error first as they surround the call to the
// Task. So, if the Task returns an error, but updating the condition to the K8s API server also
// returns an error, the K8s error will be returned. It is possible in the future errors
// are wrapped, or a slice of error is returned, but either options also bring
// a bunch of pros/cons to consider and at this time, I (P-O) just don't know which direction
// is the more user friendly.
//
// It is *required* that the Task changes the status of the Condition to its final value.
// If the condition still has the status ConditionLocked when the task returns, the
// Execute method will set the Condition to ConditionError with the Error
// set to `LockNotReleasedErr`.
func (l *Lock) Execute(ctx context.Context, task Task) (err error) {
	l.obj.Conditions().SetCondition(Condition{
		Type:   l.condition.Type,
		Status: ConditionLocked,
		Reason: "Resource locked",
	})

	if err := l.client.Status().Update(ctx, l.obj); err != nil {
		return err
	}

	err = task(l.condition)

	if err != nil {
		l.condition.Status = ConditionError
		l.condition.Reason = err.Error()
		l.obj.Conditions().SetCondition(l.condition)
	}

	if c := l.obj.Conditions().FindType(l.condition.Type); c.Status == ConditionLocked {
		l.condition.Status = ConditionError
		l.condition.Reason = LockNotReleasedErr.Error()
		l.obj.Conditions().SetCondition(l.condition)
		err = LockNotReleasedErr
	}

	if updateErr := l.client.Status().Update(ctx, l.obj); updateErr != nil {
		return updateErr
	}

	return err
}

// Returns a copy of the condition for which the lock has been created
//
// This is a helper method to allow creator of locks to easily retrieve
// the condition outside the execution loop. This can be useful if the lock
// is created but you need to check something about the condition before
// calling `Execute`
// Returns a copy of the condition for which the lock has been created
//
// This is a helper method to allow creator of locks to easily retrieve
// the condition outside the execution loop. This can be useful if the lock
// is created but you need to check something about the condition before
// calling `Execute`.
//
// This method returns a copy of the condition at the time of the creation of the
// lock.
func (l *Lock) Condition() Condition {
	return l.condition
}
