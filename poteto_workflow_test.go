package poteto

import (
	"errors"
	"testing"
)

func TestRegisterWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		workflowTypes []string
		priorities    []uint
		expecteds     []uint
	}{
		{
			"Test append workflow",
			[]string{"startUp"},
			[]uint{1},
			[]uint{1},
		},
		{
			"Test append workflow with priority",
			[]string{"startUp", "startUp"},
			[]uint{2, 1},
			[]uint{1, 2},
		},
		{
			"Test just append expected workflow",
			[]string{"unexpected", "startUp", "startUp"},
			[]uint{1, 2, 1},
			[]uint{1, 2},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			pw := &potetoWorkflows{}
			for i := range it.priorities {
				pw.RegisterWorkflow(it.workflowTypes[i], it.priorities[i], nil)
			}

			for i := range pw.startUpWorkflows {
				if pw.startUpWorkflows[i].priority != it.expecteds[i] {
					t.Errorf("Expected: %d, Got: %d", it.expecteds[i], pw.startUpWorkflows[i].priority)
				}
			}
		})
	}
}

func TestApplyStartUpWorkflows(t *testing.T) {
	tests := []struct {
		name      string
		workflows []UnitWorkflow
		hasError  bool
	}{
		{
			"Test apply start up workflows",
			[]UnitWorkflow{
				{1, func() error { return nil }},
				{2, func() error { return nil }},
			},
			false,
		},
		{
			"Test apply start up workflows with error",
			[]UnitWorkflow{
				{1, func() error { return errors.New("error") }},
				{2, func() error { return nil }},
				{3, func() error { return nil }},
			},
			true,
		},
		{
			"Test apply start up workflows with no workflows",
			[]UnitWorkflow{},
			false,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			pw := &potetoWorkflows{startUpWorkflows: it.workflows}
			err := pw.ApplyStartUpWorkflows()
			if it.hasError {
				if err == nil {
					t.Errorf("should throw an error")
				}
			} else {
				if err != nil {
					t.Errorf("should not throw an error")
				}
			}
		})
	}
}
