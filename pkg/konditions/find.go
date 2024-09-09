package konditions

// Find or initialize a condition for the type given.
// If a condition exists for the type given, it will return a *copy* of the condition
// If none exists, it will create a new condition for the type specified and the status
// will be set to ConditionInitialized
//
// Once a condition is configured and ready to be stored in a conditions, you'll
// have to add it back by calling `Conditions.SetCondition(condition)`
//
//	c := conditions.FindOrInitializeFor(ConditionType("Example"))
//	c.Status = ConditionCompleted
//	conditions.SetCondition(c)
func (c Conditions) FindOrInitializeFor(ct ConditionType) Condition {
	condition := c.FindType(ct)
	if condition != nil {
		return *condition
	}

	return Condition{
		Type:   ct,
		Status: ConditionInitialized,
	}
}

// Find the first condition that matches `ConditionStatus`
//
// This is useful,for instance, when you have a bunch of condition and you'd like
// to know if any of them has had an Error. It's important to understand that
// statuses aren't unique within a set of conditions, and as such, the condition
// returned is the first encountered.
//
//	errCondition := conditions.FindStatusCondition(api.conditionError)
//	if errCondition != nil {
//		// Log the error and mark the top level status as errored
//	}
//
// Even though a pointer is returned by the method, note that the value returned points to
// a *copy* of the condition in Conditions. This is because FindStatus can return an empty result and
// it's more explicit to return `nil` then it is to return a zered Condition.
func (c Conditions) FindStatus(conditionStatus ConditionStatus) *Condition {
	for i := range c {
		if c[i].Status == conditionStatus {
			return c[i].DeepCopy()
		}
	}

	return nil
}

// Find a condition that matches `ConditionType`.
//
// This method is similar to FindStatus but instead operates on the ConditionType. Since it is expected
// for types to be unique within a Conditions set, it should either return the same condition or nil.
//
// Even though a pointer is returned by the method, note that the value returned points to
// a *copy* of the condition in Conditions. This is because FindStatus can return an empty result and
// it's more explicit to return `nil` then it is to return a zered Condition.
func (c Conditions) FindType(conditionType ConditionType) *Condition {
	for i := range c {
		if c[i].Type == conditionType {
			return c[i].DeepCopy()
		}
	}

	return nil
}

// Check if the condition with ConditionType matches the status provided.
//
// This is an utility method that can be useful if the condition is not needed and the
// user only wants to know if a certain condition has the expected Status.
//
//	isCompleted := conditions.TypeHasStatus(ConditionType("Example"), ConditionCompleted)
//	if isCompleted {
//		// ... Do something ...
//	}
func (c Conditions) TypeHasStatus(conditionType ConditionType, status ConditionStatus) bool {
	for _, condition := range c {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}
	return false
}

// Check if any of the condition matches the ConditionStatus.
// It returns true if *any* of the conditions in the set has a status
// that matches the provided ConditionStatus.
//
//		hasError := conditions.AnyWithStatus(ConditionError)
//	 if hasError {
//			// ... Mark the object as errored ...
//		}
func (c Conditions) AnyWithStatus(status ConditionStatus) bool {
	condition := c.FindStatus(status)

	return condition != nil
}
