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

func (task *Task) CycleCheckValidatorHandler(cosmosClient *cosmosClient.Client, denom, poolAddrStr string) {
	logrus.WithFields(logrus.Fields{
		"denom":    denom,
		"poolAddr": poolAddrStr,
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
			err := task.CheckValidator(cosmosClient, denom, poolAddrStr)
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

func (task *Task) CheckValidator(cosmosClient *cosmosClient.Client, denom, poolAddrStr string) error {
	cycleSecondsRes, err := task.stafihubClient.QueryCycleSeconds(denom)
	if err != nil {
		return err
	}
	cycleInfoOnChain := cycleSecondsRes.CycleSeconds
	currentBlockHeight, err := cosmosClient.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	useSeconds := cycleInfoOnChain.Seconds

	// cal current cycle/targetHeight/slashFromHeight
	currentCycle := uint64(currentBlockHeight) / useSeconds
	targetHeight := int64(currentCycle * useSeconds)
	slashFromHeight := targetHeight - utils.SlashDuBlock

	// get local checked cycle
	localCheckedCycleVersion, localCheckedCycleNumber, found := task.getLocalCheckedCycle(denom, poolAddrStr)
	if !found {
		return fmt.Errorf("local checked cycle not exist")
	}
	// return if this cycle was already checked
	if cycleInfoOnChain.Version == localCheckedCycleVersion && currentCycle <= localCheckedCycleNumber {
		logrus.WithFields(logrus.Fields{
			"localCheckedCycleVersion": localCheckedCycleVersion,
			"localCheckedCycleNUMBER":  localCheckedCycleNumber,
			"currentCycle":             currentCycle,
		}).Debug("checkValidator no need check this cycle")
		return nil
	}
	// return if latestVotedCycle hasn't been reported

	// ---------------- check rvalidator ------------
	// 0. get rValidators on chain
	rValidatorList, err := task.stafihubClient.QueryRValidatorList(denom, poolAddrStr)
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

	// 1. collect all rValidators need rm
	needRmValidator := make([]string, 0)
	for _, validatorStr := range rValidatorList.RValidatorList {
		done := core.UseSdkConfigContext(cosmosClient.GetAccountPrefix())
		validatorAddr, err := sdk.ValAddressFromBech32(validatorStr)
		if err != nil {
			done()
			return err
		}
		done()

		// (0). rm if it has slash
		slashRes, err := cosmosClient.QueryValidatorSlashes(validatorAddr, slashFromHeight, targetHeight)
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"currentCycle": currentCycle,
			"valAddr":      validatorStr,
			"slashAmount":  slashRes.Pagination.Total,
			"fromHeight":   slashFromHeight,
			"targetHeight": targetHeight,
		}).Debug("validatorSlashInfo")

		if slashRes.Pagination.Total > utils.MaxSlashAmount {
			needRmValidator = append(needRmValidator, validatorStr)
		}

		// (1). rm if it missed blocks excessively

		// (2). rm if it's commssion is too big
		// (3). check if it is removeable(transitive redelegate is not permitted)

	}
	if len(needRmValidator) == 0 {
		logrus.Debug("needRmValidator is empty, no need redelegate")
		task.setLocalCheckedCycle(denom, poolAddrStr, cycleInfoOnChain.Version, currentCycle)
		return nil
	}

	// 2. select highquality validators from original chain, number = 2 * len(rValidatorList)
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

	//todo check transitive redelegations for old validator incase of( a -> b, b -> a )

	logrus.WithFields(logrus.Fields{
		"oldVal":        needRmValidator[0],
		"newVal":        willUseValidator[0],
		"cycleVersion:": cycleInfoOnChain.Version,
		"currentCycle":  currentCycle,
		"denom":         denom,
	}).Info("will redelegate info")

	// 3. we update one validator every cycle
	done := core.UseSdkConfigContext(stafihubClient.GetAccountPrefix())
	fromAddress := task.stafihubClient.GetFromAddress().String()
	done()

	content := stafiHubXRValidatorTypes.NewUpdateRValidatorProposal(
		fromAddress,
		denom,
		poolAddrStr,
		needRmValidator[0],
		willUseValidator[0],
		&stafiHubXRValidatorTypes.Cycle{
			Denom:   denom,
			Version: cycleInfoOnChain.Version,
			Number:  currentCycle,
		})

	err = task.checkAndReSendWithProposalContent("NewUpdateRValidatorProposal", content)
	if err != nil {
		return err
	}
	task.setLocalCheckedCycle(denom, poolAddrStr, cycleInfoOnChain.Version, currentCycle)

	return nil
}
