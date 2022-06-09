package task

import (
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/stafihub/staking-election/config"
)

type Task struct {
	stafihubClient *stafihubClient.Client
	electorAccount string
	stop           chan struct{}
}

func NewTask(cfg *config.Config, stafihubClient *stafihubClient.Client) *Task {
	s := &Task{
		stafihubClient: stafihubClient,
		electorAccount: cfg.ElectorAccount,
		stop:           make(chan struct{}),
	}
	return s
}

func (task *Task) Start() error {
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}
