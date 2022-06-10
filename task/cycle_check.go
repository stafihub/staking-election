package task

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
	"github.com/stafihub/rtoken-relay-core/common/core"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
	"github.com/stafihub/staking-election/utils"
)

var maxSlashAmount = int64(0)

func (task *Task) CycleCheckValidatorHandler(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds) {
	logrus.Infof("CycleCheckValidatorHandler start, denom: %s, cycleVersion: %d, cycleSeconds: %d",
		denom, cycleSeconds.Version, cycleSeconds.Seconds)

	ticker := time.NewTicker(time.Duration(cycleSeconds.Seconds) * time.Second)
	defer ticker.Stop()

	retry := 0
	for {
		if retry > RetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
			return
		}

		select {
		case <-task.stop:
			logrus.Info("CycleCheckValidatorHandler will stop")
			return
		case <-ticker.C:
			logrus.Info("task CycleCheckValidatorHandler start -----------")
			err := task.CheckValidator(cosmosClient, denom, cycleSeconds)
			if err != nil {
				logrus.Warnf("CheckValidator failed: %s", err)
				time.Sleep(WaitTime)
				retry++
				continue
			}

			logrus.Info("task CycleCheckValidatorHandler end -----------")
			retry = 0
		}
	}
}

func (task *Task) CheckValidator(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds) error {

	rValidatorList, err := task.stafihubClient.QueryRValidatorList(denom)
	if err != nil {
		return err
	}
	if len(rValidatorList.RValidatorList) == 0 {
		return fmt.Errorf("rValidatorList on chain is empty")
	}

	rValidatorMap := make(map[string]bool)
	for _, rval := range rValidatorList.RValidatorList {
		rValidatorMap[rval] = true
	}

	currentBlockHeight, err := cosmosClient.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	useSeconds := int64(cycleSeconds.Seconds)
	cycle := currentBlockHeight / useSeconds
	targetHeight := cycle * useSeconds

	needRmValidator := make([]string, 0)
	for _, validatorStr := range rValidatorList.RValidatorList {
		done := core.UseSdkConfigContext(cosmosClient.GetAccountPrefix())
		validatorAddr, err := sdk.ValAddressFromBech32(validatorStr)
		if err != nil {
			done()
			return err
		}
		done()

		fromHeight := targetHeight - 1000
		slashRes, err := cosmosClient.QueryValidatorSlashes(validatorAddr, fromHeight, targetHeight)
		if err != nil {
			return err
		}
		logrus.Debugf("cycle: %d, validatorSlashInfo: valAddress: %s, slashAmount: %d, fromHeight: %d, toHeight: %d",
			cycle, validatorStr, slashRes.Pagination.Total, fromHeight, targetHeight)

		// redelegate if has slash
		if slashRes.Pagination.Total >= uint64(maxSlashAmount) {
			needRmValidator = append(needRmValidator, validatorStr)
		}
	}
	if len(needRmValidator) == 0 {
		logrus.Debugf("needRmValidator is empty, no need redelegate")
		return nil
	}

	// select highquality validators, number = 2 * len(rValidatorList)
	selectedValidator, err := utils.GetSelectedValidator(cosmosClient, targetHeight, int64(len(rValidatorList.RValidatorList)*2))
	if err != nil {
		return err
	}

	willUseValidator := make([]string, 0)
	for _, val := range selectedValidator {
		if !rValidatorMap[val.OperatorAddress] {
			willUseValidator = append(willUseValidator, val.OperatorAddress)
		}
		if len(willUseValidator) == len(needRmValidator) {
			break
		}
	}

	if len(needRmValidator) != len(willUseValidator) {
		return fmt.Errorf("selected validator not enough to redelegate")
	}

	logrus.WithFields(logrus.Fields{
		"oldVal": needRmValidator[0],
		"newVal": willUseValidator[0],
	}).Info("will relegate info")

	done := core.UseSdkConfigContext(stafihubClient.GetAccountPrefix())
	fromAddress := task.stafihubClient.GetFromAddress().String()
	done()

	content := stafiHubXRValidatorTypes.NewUpdateRValidatorProposal(fromAddress, denom, needRmValidator[0], willUseValidator[0], &stafiHubXRValidatorTypes.Cycle{
		Denom:   denom,
		Version: cycleSeconds.Version,
		Number:  uint64(cycle),
	})

	return task.checkAndReSendWithProposalContent("NewUpdateRValidatorProposal", content)
}
