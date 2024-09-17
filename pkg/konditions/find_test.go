package konditions

import (
	"testing"
)

func TestFindOrInitializeFor(t *testing.T) {
	conditions := Conditions{}

	condition := conditions.FindOrInitializeFor(ConditionType("New Type"))

	if condition.Type != ConditionType("New Type") {
		t.Error("Expected a valid condition with the same type as defined")
	}

	if condition.Status != ConditionInitialized {
		t.Error("Expected a new condition to have the status set to Initialized")
	}

	err := conditions.SetCondition(Condition{
		Type:   ConditionType("Existing"),
		Status: ConditionCompleted,
	})

	if err != nil {
		t.Error("Expected SetCondition to add the condition")
	}

	condition = conditions.FindOrInitializeFor(ConditionType("Existing"))
	if condition.Status != ConditionCompleted {
		t.Error("Expected the condition returned to be completed")
	}
}

func TestFindStatus(t *testing.T) {
	conditions := Conditions{
		{
			Status: ConditionLocked,
			Type:   ConditionType("locked condition"),
		},
		{
			Status: ConditionInitialized,
			Type:   ConditionType("initialized condition"),
		},
		{
			Status: ConditionLocked,
			Type:   ConditionType("locked condition #2"),
		},
		{
			Status: ConditionCompleted,
			Type:   ConditionType("completed condition"),
		},
	}

	if condition := conditions.FindStatus(ConditionLocked); condition == nil || condition.Type != ConditionType("locked condition") {
		t.Error("Unexpected condition: ", condition)
	}

	if condition := conditions.FindStatus(ConditionTerminated); condition != nil {
		t.Error("Expected to find no condition, found: ", condition)
	}

	if condition := conditions.FindStatus(ConditionInitialized); condition == nil || condition.Type != ConditionType("initialized condition") {
		t.Error("Unexpected condition: ", condition)
	}
}

func TestFindType(t *testing.T) {
	authorizationType := ConditionType("authorization")
	connectingType := ConditionType("connecting")
	conditions := Conditions{
		{
			Status: ConditionLocked,
			Type:   authorizationType,
		},
		{
			Status: ConditionInitialized,
			Type:   connectingType,
		},
		{
			Status: ConditionLocked,
			Type:   ConditionType("workspace bridge"),
		},
		{
			Status: ConditionCompleted,
			Type:   ConditionType("component configured"),
		},
	}

	if condition := conditions.FindType(authorizationType); condition == nil || condition.Type != authorizationType {
		t.Error("Unexpected condition: ", condition)
	}

	if condition := conditions.FindType(connectingType); condition == nil || condition.Type != connectingType {
		t.Error("Unexpected condition: ", condition)
	}

	if condition := conditions.FindType(ConditionType("Non-existing Type")); condition != nil {
		t.Error("Unexpected condition: ", condition)
	}
}

func TestTypeHasStatus(t *testing.T) {
	if result := (Conditions{}).TypeHasStatus(ConditionType("Any Type on empty conditions"), ConditionLocked); result != false {
		t.Error("Empty conditions should have returned false")
	}

	conditions := Conditions{
		{
			Type:   ConditionType("Negotiating"),
			Status: ConditionCompleted,
		},
		{
			Type:   ConditionType("Signing"),
			Status: ConditionError,
		},
		{
			Type:   ConditionType("Sent"),
			Status: ConditionInitialized,
		},
	}

	if result := conditions.TypeHasStatus(ConditionType("Negotiating"), ConditionCompleted); result == false {
		t.Error("Expected to return true")
	}

	if result := conditions.TypeHasStatus(ConditionType("Signing"), ConditionError); result == false {
		t.Error("Expected to return true")
	}

	if result := conditions.TypeHasStatus(ConditionType("Signing"), ConditionCompleted); result == true {
		t.Error("Expected to return false")
	}

	if result := conditions.TypeHasStatus(ConditionType("Sent"), ConditionCompleted); result == true {
		t.Error("Expected to return false")
	}

	if result := conditions.TypeHasStatus(ConditionType("Non Existant"), ConditionCompleted); result == true {
		t.Error("Expected to return false")
	}
}

func TestAnyWithStatus(t *testing.T) {
	if result := (Conditions{}).AnyWithStatus(ConditionLocked); result != false {
		t.Error("Empty conditions should have returned false")
	}

	conditions := Conditions{
		{
			Type:   ConditionType("Negotiating"),
			Status: ConditionCompleted,
		},
		{
			Type:   ConditionType("Signing"),
			Status: ConditionError,
		},
		{
			Type:   ConditionType("Sent"),
			Status: ConditionInitialized,
		},
	}

	if result := conditions.AnyWithStatus(ConditionCompleted); result == false {
		t.Error("Expected to return true")
	}

	if result := conditions.AnyWithStatus(ConditionError); result == false {
		t.Error("Expected to return true")
	}

	if result := conditions.AnyWithStatus(ConditionInitialized); result == false {
		t.Error("Expected to return true")
	}

	if result := conditions.AnyWithStatus(ConditionLocked); result == true {
		t.Error("Expected to return false")
	}
}
