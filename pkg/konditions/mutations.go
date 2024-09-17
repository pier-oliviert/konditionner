package konditions

import (
	"errors"
	"slices"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var NotInitializedConditionsErr = errors.New("Conditions is not initialized")

// Set the given condition into the Conditions.
// The return value indicates whether the condition was changed in the stack or not.
//
// This is the main method you'll use on a Conditions set to add/update a condition that
// you have operated on. The condition will be stored in the set but won't be persisted
// until you actually run the update/patch command to the Kubernetes server.
//
//	myNewCondition := Condition{
//		Type: ConditionType("A Controlled Step"),
//		Status: ConditionCreated,
//		Reason: "Item Created, waiting until it becomes available",
//	}
//	myResource.conditions.SetCondition(myNewCondition)
//	if err := reconciler.Status().Update(&myResource); err != nil {
//		// ... deal with k8s error ...
//	}
func (c *Conditions) SetCondition(newCondition Condition) error {
	if c == nil {
		return NotInitializedConditionsErr
	}

	if newCondition.LastTransitionTime.IsZero() {
		newCondition.LastTransitionTime = meta.NewTime(time.Now())
	}

	var condition *Condition
	var index int
	for i, _ := range *c {
		existing := &((*c)[i])
		if existing.Type == newCondition.Type {
			condition = existing
			index = i
		}
	}

	if condition == nil {
		*c = append(*c, newCondition)
		return nil
	}

	*c = slices.Replace(*c, index, index+1, newCondition)
	return nil
}

// Remove the conditionType from the conditions set.
// The return value indicates whether a condition was removed or not.
//
// Since all conditions are identified by a ConditionType, only the type is needed
// when removing a condition from the set. The changes won't be persisted until
// you actually run the update/patch command to the Kubernetes server.
//
//	myResource.conditions.RemoveConditionWith(ConditionType("A Controller Step"))
//	if err := reconciler.Status().Update(&myResource); err != nil {
//		// ... deal with k8s error ...
//	}
func (c *Conditions) RemoveConditionWith(conditionType ConditionType) (removed bool) {
	if c == nil || len(*c) == 0 {
		return false
	}
	newConditions := make(Conditions, 0, len(*c)-1)
	for _, condition := range *c {
		if condition.Type != conditionType {
			newConditions = append(newConditions, condition)
		}
	}

	removed = len(*c) != len(newConditions)
	*c = newConditions

	return removed
}
