package konditions

import (
	"testing"
)

func TestConditionStatusIsOneOf(t *testing.T) {
	condition := Condition{
		Type:   ConditionType("example"),
		Status: ConditionCompleted,
	}

	if condition.StatusIsOneOf(ConditionLocked) == true {
		t.Error("Status is not Locked but returned true")
	}

	if condition.StatusIsOneOf(ConditionLocked, ConditionTerminated) == true {
		t.Error("Status is not Locked, nor is it terminated. It returned true")
	}

	if condition.StatusIsOneOf(ConditionLocked, ConditionCompleted) == false {
		t.Error("Status is Completed, should return true")
	}

	if condition.StatusIsOneOf(ConditionCompleted) == false {
		t.Error("Status is Completed, should return true")
	}

	if condition.StatusIsOneOf(ConditionTerminated, ConditionTerminating, ConditionCompleted) == false {
		t.Error("Status is Completed, should return true")
	}
}

func TestConditionsDeepCopy(t *testing.T) {
	length := 3

	conditions := make(Conditions, length)

	newConditions := conditions.DeepCopy()
	newConditions = append(newConditions, Condition{})

	if len(conditions) != 3 {
		t.Fail()
	}

	if len(newConditions) != 4 {
		t.Fail()
	}
}

func TestConditionDeepCopy(t *testing.T) {
	c := Condition{
		Type:   ConditionType("test"),
		Status: ConditionInitialized,
		Reason: "Go!",
	}

	newCond := c.DeepCopy()
	newCond.Status = ConditionCompleted
	newCond.Reason = "Finished!"

	if c.Status == newCond.Status {
		t.Error("Status shouldn't be the same")
	}

	if c.Reason == newCond.Reason {
		t.Error("Status shouldn't be the same")
	}
}
