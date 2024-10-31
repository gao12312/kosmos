package workflow

import (
	"context"
	"time"

	"github.com/kosmos.io/kosmos/pkg/apis/kosmos/v1alpha1"
	"github.com/kosmos.io/kosmos/pkg/kubenest/controller/virtualcluster.node.controller/workflow/task"
)

const (
	retryCount = 0
	maxRetries = 5
)

// nolint:revive
type WorkflowData struct {
	Tasks []task.Task
}

func RunWithRetry(ctx context.Context, task task.Task, opt task.TaskOpt, preArgs interface{}) (interface{}, error) {
	i := retryCount
	var err error
	var args interface{}
	for ; i < maxRetries; i++ {
		if args, err = task.Run(ctx, opt, preArgs); err != nil {
			if !task.Retry {
				break
			}
			waitTime := 3 * (i + 1)
			opt.Loger().Infof("work flow retry %d after %ds, task name: %s, err: %s", i, waitTime, task.Name, err)
			time.Sleep(time.Duration(waitTime) * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		if task.ErrorIgnore {
			opt.Loger().Infof("work flow ignore err, task name: %s, err: %s", task.Name, err)
			return nil, nil
		}
		opt.Loger().Infof("work flow interrupt, task name: %s, err: %s", task.Name, err)
		return nil, err
	}
	return args, nil
}

// nolint:revive
func (w WorkflowData) RunTask(ctx context.Context, opt task.TaskOpt) error {
	var args interface{}
	for i, t := range w.Tasks {
		opt.Loger().Infof("HHHHHHHHHHHH (%d/%d) work flow run task %s  HHHHHHHHHHHH", i+1, len(w.Tasks), t.Name)
		if t.Skip != nil && t.Skip(ctx, opt) {
			opt.Loger().Infof("work flow skip task %s", t.Name)
			continue
		}
		if len(t.SubTasks) > 0 {
			for j, subTask := range t.SubTasks {
				opt.Loger().Infof("HHHHHHHHHHHH (%d/%d) work flow run sub task %s HHHHHHHHHHHH", j+1, len(t.SubTasks), subTask.Name)
				if t.Skip != nil && t.Skip(ctx, opt) {
					opt.Loger().Infof("work flow skip sub task %s", t.Name)
					continue
				}

				if nextArgs, err := RunWithRetry(ctx, subTask, opt, args); err != nil {
					return err
				} else {
					args = nextArgs
				}
			}
		} else {
			if nextArgs, err := RunWithRetry(ctx, t, opt, args); err != nil {
				return err
			} else {
				args = nextArgs
			}
		}
	}
	return nil
}

func NewJoinWorkFlow() WorkflowData {
	joinTasks := []task.Task{
		task.NewCheckEnvTask(),
		task.NewDrainHostNodeTask(),
		task.NewKubeadmResetTask(),
		task.NewCleanHostClusterNodeTask(),
		task.NewReomteUploadCATask(),
		task.NewRemoteUpdateKubeletConfTask(),
		task.NewRemoteUpdateConfigYamlTask(),
		task.NewRemoteNodeJoinTask(),
		task.NewWaitNodeReadyTask(false),
		task.NewInstallLBTask(),
		task.NewUpdateVirtualNodeLabelsTask(),
		task.NewUpdateNodePoolItemStatusTask(v1alpha1.NodeInUse, false),
	}

	return WorkflowData{
		Tasks: joinTasks,
	}
}

func NewUnjoinWorkFlow() WorkflowData {
	unjoinTasks := []task.Task{
		task.NewCheckEnvTask(),
		task.NewDrainVirtualNodeTask(),
		task.NewRemoveNodeFromVirtualTask(),
		task.NewExecShellUnjoinCmdTask(),
		task.NewJoinNodeToHostCmd(),
		task.NewUpdateHostNodeLabelsTask(),
		task.NewUpdateNodePoolItemStatusTask(v1alpha1.NodeFreeState, true),
	}
	return WorkflowData{
		Tasks: unjoinTasks,
	}
}

func NewCleanNodeWorkFlow() WorkflowData {
	cleanNodeTasks := []task.Task{
		task.NewUpdateNodePoolItemStatusTask(v1alpha1.NodeFreeState, true),
	}
	return WorkflowData{
		Tasks: cleanNodeTasks,
	}
}
