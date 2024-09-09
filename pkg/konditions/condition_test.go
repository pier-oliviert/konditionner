package konditions

import (
	"testing"
)

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
