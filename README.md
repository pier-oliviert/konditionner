# Konditionner: Conditions for K8s Custom Resources
[![Main](https://github.com/pier-oliviert/konditionner/actions/workflows/main.yaml/badge.svg)](https://github.com/pier-oliviert/konditionner/actions/workflows/main.yaml)

This library exists to help manage conditions for people that builds Custom Resources Definitions(CRDs) for Kubernetes. Konditionner is built on the same idea as the utility package from [API Machinery](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition) but with extensibility in mind.

The library is a great addition to your custom resources if you're operator uses [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).

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

    if record.Conditions().AnyWithStatus(konditions.ConditionError) {
        // The record has failed, do not touch the record
        return ctrl.Result{}, nil
    }

    lock := konditions.NewLock(record, r.Client, VolumeFinder)
    err := lock.Execute(ctx, func(condition konditions.Condition) error {
        // Volume Finder need to be reconciled.
        err := r.ReconcileVolume(ctx, record, &condition)
        if err != nil {
            return err
        }

        record.Conditions().SetCondition(condition)
        return ctrl.Result{}, r.Status().Update(ctx, record)
    })

    // ...

    return ctrl.Result{}, err
}

func (r *MyRecordReconciler) ReconcileVolume(ctx context.Context, record *MyRecord, condition *konditions.Condition) error {
    switch condition.Status {
    case konditions.ConditionInitialized:
        // ... Start working on the Volume Finder condition

        if err != nil {
            return err
        }

        condition.Status = konditions.ConditionCreated
        condition.Reason = "Volume Found"
        return nil

    case konditions.ConditionCreated:
        // ... Volume was created, let's configure it for our record
    }

    return nil
}

```



## Contributing

If you'd like to help or you need to make modification to the project to make it better for you, feel free to create issues/pull requests as you see fit. I don't really know the scope of this project yet, so I'll be pretty flexible as to what can be part of Konditionner.
