package konditions

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
func (c *Conditions) SetCondition(newCondition Condition) (changed bool) {
	if c == nil {
		return false
	}

	existingCondition := c.FindType(newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = meta.NewTime(time.Now())
		}
		*c = append(*c, newCondition)
		return true
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		if !newCondition.LastTransitionTime.IsZero() {
			existingCondition.LastTransitionTime = newCondition.LastTransitionTime
		} else {
			existingCondition.LastTransitionTime = meta.NewTime(time.Now())
		}
		changed = true
	}

	if existingCondition.Reason != newCondition.Reason {
		existingCondition.Reason = newCondition.Reason
		changed = true
	}

	return changed
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
