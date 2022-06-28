package task

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sirupsen/logrus"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	stafiHubXRelayersTypes "github.com/stafihub/stafihub/x/relayers/types"
	stafiHubXRVoteTypes "github.com/stafihub/stafihub/x/rvote/types"
	"github.com/stafihub/staking-election/config"
	"github.com/stafihub/staking-election/utils"
)

var RetryLimit = 100
var WaitTime = time.Second * 6
var cycleFactor = uint64(1e10)

type Task struct {
	stafihubClient       *stafihubClient.Client
	electorAccount       string
	stafihubEndpointList []string
	rTokenInfoMap        map[string]config.RTokenInfo
	localCheckedCycle    sync.Map // avoid repeated check
	stop                 chan struct{}
}

func NewTask(cfg *config.Config, stafihubClient *stafihubClient.Client) *Task {
	rTokenInfoMap := make(map[string]config.RTokenInfo)
	for _, rtokenInfo := range cfg.RTokenInfo {
		rTokenInfoMap[rtokenInfo.Denom] = rtokenInfo
	}

	s := &Task{
		stafihubClient:       stafihubClient,
		electorAccount:       cfg.ElectorAccount,
		stafihubEndpointList: cfg.StafiHubEndpointList,
		rTokenInfoMap:        rTokenInfoMap,
		stop:                 make(chan struct{}),
	}
	return s
}

func (task *Task) setLocalCheckedCycle(denom, poolAddrStr string, cycleVersion, cycleNumber uint64) {
	task.localCheckedCycle.Store(denom+poolAddrStr, cycleVersion*cycleFactor+cycleNumber)
}

func (task *Task) getLocalCheckedCycle(denom, poolAddrStr string) (cycleVersion, cycleNumber uint64, found bool) {
	anyValue, found := task.localCheckedCycle.Load(denom + poolAddrStr)
	if !found {
		return 0, 0, false
	}
	value := anyValue.(uint64)
	return value / cycleFactor, value % cycleFactor, true
}

func (task *Task) Start() error {
	for _, rTokenInfo := range task.rTokenInfoMap {

		addressPrefixRes, err := task.stafihubClient.QueryAddressPrefix(rTokenInfo.Denom)
		if err != nil {
			return err
		}
		cosmosClient, err := cosmosClient.NewClient(nil, "", "", addressPrefixRes.AccAddressPrefix, rTokenInfo.EndpointList)
		if err != nil {
			return err
		}

		bondedPoolsRes, err := task.stafihubClient.QueryPools(rTokenInfo.Denom)
		if err != nil {
			return err
		}

		for _, poolAddrStr := range bondedPoolsRes.Addrs {
			cycle, err := task.stafihubClient.QueryLatestVotedCycle(rTokenInfo.Denom, poolAddrStr)
			if err != nil {
				return err
			}
			task.setLocalCheckedCycle(rTokenInfo.Denom, poolAddrStr, cycle.LatestVotedCycle.Version, cycle.LatestVotedCycle.Number)

			utils.SafeGoWithRestart(func() {
				task.CycleCheckValidatorHandler(cosmosClient, rTokenInfo.Denom, poolAddrStr)
			})
		}
	}

	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (h *Task) checkAndReSendWithProposalContent(typeStr string, content stafiHubXRVoteTypes.Content) error {
	logrus.WithFields(logrus.Fields{
		"type": typeStr,
	}).Info("checkAndReSendWithProposalContent start")

	txHashStr, _, err := h.stafihubClient.SubmitProposal(content)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), stafiHubXRelayersTypes.ErrAlreadyVoted.Error()):
			logrus.WithFields(logrus.Fields{
				"type": typeStr,
			}).Info("no need send, already voted")
			return nil
		case strings.Contains(err.Error(), stafiHubXRVoteTypes.ErrProposalAlreadyApproved.Error()):
			logrus.WithFields(logrus.Fields{
				"type": typeStr,
			}).Info("no need send, already approved")
			return nil
		case strings.Contains(err.Error(), stafiHubXRVoteTypes.ErrProposalAlreadyExpired.Error()):
			logrus.WithFields(logrus.Fields{
				"type": typeStr,
			}).Info("no need send, already expired")
			return nil

		// resend case:
		case strings.Contains(err.Error(), errors.ErrWrongSequence.Error()):
			return h.checkAndReSendWithProposalContent(txHashStr, content)
		}

		return err
	}
	logrus.WithFields(logrus.Fields{
		"txhash":  txHashStr,
		"typeStr": typeStr,
	}).Debug("checkAndReSendWithProposalContent")

	retry := RetryLimit
	var res *sdk.TxResponse
	for {
		if retry <= 0 {
			logrus.WithFields(logrus.Fields{
				"tx hash": txHashStr,
				"err":     err,
			}).Error("checkAndReSendWithProposalContent QueryTxByHash, reach retry limit.")
			return fmt.Errorf("checkAndReSendWithProposalContent QueryTxByHash reach retry limit, tx hash: %s,err: %s", txHashStr, err)
		}

		//check on chain
		res, err = h.stafihubClient.QueryTxByHash(txHashStr)
		if err != nil || res.Empty() || res.Height == 0 {
			if res != nil {
				logrus.Debug(fmt.Sprintf(
					"checkAndReSendWithProposalContent QueryTxByHash, tx failed. will query after %f second",
					WaitTime.Seconds()),
					"tx hash", txHashStr,
					"res.log", res.RawLog,
					"res.code", res.Code)
			} else {
				logrus.Debug(fmt.Sprintf(
					"checkAndReSendWithProposalContent QueryTxByHash failed. will query after %f second",
					WaitTime.Seconds()),
					"tx hash", txHashStr,
					"err", err)
			}

			time.Sleep(WaitTime)
			retry--
			continue
		}

		if res.Code != 0 {
			switch {
			case strings.Contains(res.RawLog, stafiHubXRelayersTypes.ErrAlreadyVoted.Error()):
				logrus.WithFields(logrus.Fields{
					"txHash": txHashStr,
					"type":   typeStr,
				}).Info("no need send, already voted")
				return nil
			case strings.Contains(res.RawLog, stafiHubXRVoteTypes.ErrProposalAlreadyApproved.Error()):
				logrus.WithFields(logrus.Fields{
					"txHash": txHashStr,
					"type":   typeStr,
				}).Info("no need send, already approved")
				return nil
			case strings.Contains(res.RawLog, stafiHubXRVoteTypes.ErrProposalAlreadyExpired.Error()):
				logrus.WithFields(logrus.Fields{
					"txHash": txHashStr,
					"type":   typeStr,
				}).Info("no need send, already expired")
				return nil

			// resend case
			case strings.Contains(res.RawLog, errors.ErrOutOfGas.Error()):
				return h.checkAndReSendWithProposalContent(txHashStr, content)
			default:
				return fmt.Errorf("tx failed, txHash: %s, rawlog: %s", txHashStr, res.RawLog)
			}
		}

		break
	}

	logrus.WithFields(logrus.Fields{
		"txHash": txHashStr,
		"type":   typeStr,
	}).Info("checkAndReSendWithProposalContent success")
	return nil
}
