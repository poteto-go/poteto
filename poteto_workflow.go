package poteto

import (
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/tslice"
)

type UnitWorkflow struct {
	priority uint
	workflow WorkflowFunc
}

// workflow is a function that is executed when the server starts | end
// - constant.START_UP_WORKFLOW: "startUp"
//   - This is a workflow that is executed when the server starts
type PotetoWorkflows interface {
	RegisterWorkflow(workflowType string, priority uint, workflow WorkflowFunc)
	ApplyStartUpWorkflows() error
}

type potetoWorkflows struct {
	startUpWorkflows []UnitWorkflow
}

func NewPotetoWorkflows() PotetoWorkflows {
	return &potetoWorkflows{
		startUpWorkflows: []UnitWorkflow{},
	}
}

func (pw *potetoWorkflows) RegisterWorkflow(workflowType string, priority uint, workflow WorkflowFunc) {
	switch workflowType {
	case constant.START_UP_WORKFLOW:
		pw.startUpWorkflows = append(pw.startUpWorkflows, UnitWorkflow{priority, workflow})
		pw.startUpWorkflows = sortWorkflows(pw.startUpWorkflows)
	default:
		// pass
	}
}

func (pw *potetoWorkflows) ApplyStartUpWorkflows() error {
	if len(pw.startUpWorkflows) == 0 {
		return nil
	}

	for _, workflow := range pw.startUpWorkflows {
		if err := workflow.workflow(); err != nil {
			return err
		}
	}
	return nil
}

func sortWorkflows(workflows []UnitWorkflow) []UnitWorkflow {
	tslice.Sort(workflows, func(l, r UnitWorkflow) int {
		return int(l.priority) - int(r.priority)
	})
	return workflows
}
