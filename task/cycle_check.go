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

func (task *Task) CycleCheckValidatorHandler(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds) {
	logrus.WithFields(logrus.Fields{
		"denom":        denom,
		"cycleVersion": cycleSeconds.Version,
		"cycleSeconds": cycleSeconds.Seconds,
	}).Info("CycleCheckValidatorHandler start")

	ticker := time.NewTicker(time.Duration(60) * time.Second)
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
			logrus.Info("task CycleCheckValidatorHandler start ----------->")
			err := task.CheckValidator(cosmosClient, denom, cycleSeconds)
			if err != nil {
				logrus.Warnf("CheckValidator failed: %s", err)
				time.Sleep(WaitTime)
				retry++
				continue
			}

			logrus.Info("task CycleCheckValidatorHandler end <-----------")
			retry = 0
		}
	}
}

func (task *Task) CheckValidator(cosmosClient *cosmosClient.Client, denom string, cycleSeconds *stafiHubXRValidatorTypes.CycleSeconds) error {
	currentBlockHeight, err := cosmosClient.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	useSeconds := cycleSeconds.Seconds
	cycle := uint64(currentBlockHeight) / useSeconds
	targetHeight := int64(cycle * useSeconds)
	slashFromHeight := targetHeight - utils.SlashDuBlock

	dealedCycleAny, ok := task.dealedCycle.Load(denom)
	if !ok {
		return fmt.Errorf("dealed Cycle not exist")
	}
	dealedCycle := dealedCycleAny.(uint64)
	if cycle <= dealedCycle {
		logrus.WithFields(logrus.Fields{
			"dealedCycle": dealedCycle,
			"cycle":       cycle,
		}).Debug("checkValidator no need deal")
		return nil
	}

	// get on chain rValidators
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

	// collector all rValidators need rm
	needRmValidator := make([]string, 0)
	for _, validatorStr := range rValidatorList.RValidatorList {
		done := core.UseSdkConfigContext(cosmosClient.GetAccountPrefix())
		validatorAddr, err := sdk.ValAddressFromBech32(validatorStr)
		if err != nil {
			done()
			return err
		}
		done()

		slashRes, err := cosmosClient.QueryValidatorSlashes(validatorAddr, slashFromHeight, targetHeight)
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"cycle":        cycle,
			"valAddr":      validatorStr,
			"slashAmount":  slashRes.Pagination.Total,
			"fromHeight":   slashFromHeight,
			"targetHeight": targetHeight,
		}).Debug("validatorSlashInfo")

		// redelegate if has slash
		if slashRes.Pagination.Total >= utils.MaxSlashAmount {
			needRmValidator = append(needRmValidator, validatorStr)
		}
	}
	if len(needRmValidator) == 0 {
		logrus.Debug("needRmValidator is empty, no need redelegate")
		task.dealedCycle.Store(denom, cycle)
		return nil
	}

	// select highquality validators from original chain, number = 2 * len(rValidatorList)
	selectedValidator, err := utils.GetSelectedValidator(cosmosClient, targetHeight, int64(len(rValidatorList.RValidatorList)*3), nil)
	if err != nil {
		return err
	}

	willUseValidator := make([]string, 0)
	for _, val := range selectedValidator {

		done := core.UseSdkConfigContext(cosmosClient.GetAccountPrefix())
		valAddress, err := sdk.ValAddressFromBech32(val.OperatorAddress)
		if err != nil {
			return err
		}
		done()

		//should remove slashed validators
		slashRes, err := cosmosClient.QueryValidatorSlashes(valAddress, slashFromHeight, targetHeight)
		if err != nil {
			return err
		}
		if slashRes.Pagination.Total > utils.MaxSlashAmount {
			continue
		}

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

	// checkout redelegations with old validator incase of( a -> b, b -> a )

	logrus.WithFields(logrus.Fields{
		"oldVal":        needRmValidator[0],
		"newVal":        willUseValidator[0],
		"cycleVersion:": cycleSeconds.Version,
		"cycle":         cycle,
		"denom":         denom,
	}).Info("will redelegate info")

	// we update one validator every cycle
	done := core.UseSdkConfigContext(stafihubClient.GetAccountPrefix())
	fromAddress := task.stafihubClient.GetFromAddress().String()
	done()

	content := stafiHubXRValidatorTypes.NewUpdateRValidatorProposal(fromAddress, denom, needRmValidator[0], willUseValidator[0], &stafiHubXRValidatorTypes.Cycle{
		Denom:   denom,
		Version: cycleSeconds.Version,
		Number:  cycle,
	})

	err = task.checkAndReSendWithProposalContent("NewUpdateRValidatorProposal", content)
	if err != nil {
		return err
	}
	task.dealedCycle.Store(denom, cycle)

	return nil
}
