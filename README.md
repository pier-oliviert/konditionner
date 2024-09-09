# Konditionner: Conditions for K8s Custom Resources

This library exists to help manage conditions for people that builds Custom Resources Definitions(CRDs) for Kubernetes.

## Features

- Reasonable default values for you to use;
- Extensible Status and Types as you see fit;
- Set of utility functions to create, update and delete conditions;

## Works natively with Kubernetes Custom Resources

Conditions are ideally used with [Status subresources](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#subresources):

```go
# A Custom Resource Spec & Status

type MyRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySpec   `json:"spec,omitempty"`
	Status MyStatus `json:"status,omitempty"`
}

type MySpec struct {
 // ... Fields ...
}

type VolumeFinder konditions.ConditionType = "Volumes"
type SecretMapping konditions.ConditionType = "Secrets"

type MyStatus struct {
    Conditions konditions.Conditions `json:"conditions,omitempty"`
}

```

Once the record includes conditions, you can use it to control the handling of your resource within the reconciliation loop.

```go
func (r *MyRecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    var record MyRecord
    if err := r.Get(ctx, req, &record); err != nil {
        return ctrl.Result{}, err
    }

    if record.Status.Conditions.AnyWithStatus(konditions.ConditionError) {
        // The record has failed, do not touch the record
        return ctrl.Result{}, nil
    }

    if !record.Status.Conditions.TypeHasStatus(VolumeFinder, konditions.ConditionCompleted) {
        // Volume Finder need to be reconciled.
        return r.ReconcileVolume(ctx, record)
    }

    // ...
}

func (r *MyRecordReconciler) ReconcileVolume(ctx context.Context, record *MyRecord) (ctrl.Result, error) {
    // FindOrInitializeFor returns a copy so we can make changes to the condition as needed.
    condition := record.Status.Conditions.FindOrInitializeFor(VolumeFinder)
    
    switch condition.Status {
    case konditions.ConditionInitialized:
        // ... Start working on the Volume Finder condition

        if err != nil {
            condition.Status = konditions.ConditionError
            condition.Reason = err.Error()

            // Since we're working with a copy, the condition needs to be added back to the
            // set. Then, an update will need to occur to persist the changes.
            record.Status.Conditions.SetCondition(condition)
        }

    case konditions.ConditionCreated:
        // ... Volume was created, let's configure it for our record
    }

    return ctrl.Result{}, r.Status().Update(ctx, record)
}

```



## Contributing

If you'd like to help or you need to make modification to the project to make it better for you, feel free to create issues/pull requests as you see fit. I don't really know the scope of this project yet, so I'll be pretty flexible as to what can be part of Konditionner.
