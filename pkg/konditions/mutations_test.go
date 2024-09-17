package konditions

import (
	"testing"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetConditionOnNil(t *testing.T) {
	var conditions *Conditions

	err := conditions.SetCondition(Condition{
		Type:   ConditionType("New Type"),
		Status: ConditionTerminated,
	})

	if err == nil {
		t.Error("Conditions not initialized yet, shouldn't be able to add values")
	}
}

func TestSetCondition(t *testing.T) {
	status := struct {
		conditions Conditions
	}{
		conditions: Conditions{},
	}

	err := status.conditions.SetCondition(Condition{
		Type:   ConditionType("New Type"),
		Status: ConditionTerminated,
	})

	if err != nil {
		t.Error("The condition should have been added now that the pointer is set")
	}

	condition := status.conditions.FindType(ConditionType("New Type"))
	if condition == nil {
		t.Error("Expected to find a condition after setting it")
	}

	if condition.Type != ConditionType("New Type") {
		t.Error("Expected the condition type to match")
	}

	if condition.Status != ConditionTerminated {
		t.Error("Expected the condition status to match")
	}

	if condition.LastTransitionTime.IsZero() {
		t.Error("Expected LastTransitionTime to be set")
	}

	// Add 2 more conditions for the next test step
	status.conditions.SetCondition(Condition{
		Type:   ConditionType("ToBeReplaced"),
		Status: ConditionTerminated,
	})

	status.conditions.SetCondition(Condition{
		Type:   ConditionType("3rd Type"),
		Status: ConditionTerminated,
	})

	if len(status.conditions) != 3 {
		t.Error("Expected 3 conditions")
	}

	status.conditions.SetCondition(Condition{
		Type:   ConditionType("ToBeReplaced"),
		Status: ConditionCompleted,
		Reason: "Replaced!",
	})

	found := status.conditions.FindType(ConditionType("ToBeReplaced"))
	if found.Status != ConditionCompleted {
		t.Error("Expected condition to be completed")
	}

	if found.Reason != "Replaced!" {
		t.Error("Expected condition to be replaced")
	}

	if len(status.conditions) != 3 {
		t.Errorf("Expected 3 conditions, had %d", len(status.conditions))
	}

	if status.conditions.FindType(ConditionType("3rd Type")) == nil {
		t.Error("Missing condition")
	}

	if status.conditions.FindType(ConditionType("New Type")) == nil {
		t.Error("Missing condition")
	}
}

func TestRemoveCondition(t *testing.T) {
	var conditions *Conditions

	removed := conditions.RemoveConditionWith(ConditionType("Test"))
	if removed == true {
		t.Error("Conditions not initialized, should have not removed anything")
	}

	conditions = &Conditions{}

	removed = conditions.RemoveConditionWith(ConditionType("Test"))
	if removed == true {
		t.Error("Conditions empty, should have not removed anything")
	}

	*conditions = append(*conditions, Condition{
		Type:               ConditionType("Test"),
		Status:             ConditionError,
		LastTransitionTime: meta.NewTime(time.Now()),
	}, Condition{
		Type:               ConditionType("Do Not Remove"),
		Status:             ConditionCompleted,
		LastTransitionTime: meta.NewTime(time.Now()),
	}, Condition{
		Type:               ConditionType("Remove This One"),
		Status:             ConditionError,
		LastTransitionTime: meta.NewTime(time.Now()),
	})

	removed = conditions.RemoveConditionWith(ConditionType("Remove This One"))

	if removed == false {
		t.Error("Conditions should have been removed")
	}

	if condition := conditions.FindType(ConditionType("Remove This One")); condition != nil {
		t.Error("Expected the condition to be removed, got: ", condition)
	}

	if condition := conditions.FindType(ConditionType("Test")); condition == nil {
		t.Error("Expected the condition to still be present")
	}

	if condition := conditions.FindType(ConditionType("Do Not Remove")); condition == nil {
		t.Error("Expected the condition to still be present")
	}
}
