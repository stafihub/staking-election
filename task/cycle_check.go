package task

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
	"github.com/stafihub/rtoken-relay-core/common/core"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
	"github.com/stafihub/staking-election/utils"
)

func (task *Task) CycleCheckValidatorHandler(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds, validatorNumber int64) {

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
			err := task.CheckValidator(cosmosClient, denom, cycleSeconds, validatorNumber)
			if err != nil {
				time.Sleep(WaitTime)
				retry++
				continue
			}

			logrus.Info("task CycleCheckValidatorHandler end -----------")
			retry = 0
		}
	}
}

func (task *Task) CheckValidator(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds, validatorNumber int64) error {

	rValidatorList, err := task.stafihubClient.QueryRValidatorList(denom)
	if err != nil {
		return err
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

		slashRes, err := cosmosClient.QueryValidatorSlashes(validatorAddr, targetHeight-1000, targetHeight)
		if err != nil {
			return err
		}

		// redelegate if has slash
		if slashRes.Pagination.Total > 2 {
			needRmValidator = append(needRmValidator, validatorStr)

		}
	}
	if len(needRmValidator) == 0 {
		return nil
	}
	selectedValidator, err := utils.GetSelectedValidator(cosmosClient, targetHeight, validatorNumber*2)
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

	content := stafiHubXRValidatorTypes.NewUpdateRValidatorProposal(task.stafihubClient.GetFromName(), denom, needRmValidator[0], willUseValidator[0])

	return task.checkAndReSendWithProposalContent("NewUpdateRValidatorProposal", content)
}
