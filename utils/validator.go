package utils

import (
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
)

var averageBlockTIme = sdk.MustNewDecFromStr("6.77")

func GetAverageAnnualRate(c *cosmosClient.Client, height int64) (sdk.Dec, error) {
	valMap, err := GetValidatorAnnualRate(c, height)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	totalAnuualRate := sdk.NewDec(0)
	for _, val := range valMap {
		totalAnuualRate = totalAnuualRate.Add(val.AnnualRate)
	}

	initialLen := len(valMap)

	return totalAnuualRate.Quo(sdk.NewDec(int64(initialLen))), nil
}

func GetSelectedValidator(c *cosmosClient.Client, height, number int64) ([]*Validator, error) {
	valMap, err := GetValidatorAnnualRate(c, height)
	if err != nil {
		return nil, err
	}

	valSlice := make([]*Validator, 0)
	for _, val := range valMap {
		valSlice = append(valSlice, val)
	}
	sort.Slice(valSlice, func(i, j int) bool {
		return valSlice[i].TokenAmount.GT(valSlice[j].TokenAmount)
	})

	initialLen := len(valSlice)
	// return all validators if not enough
	if initialLen <= int(number) {
		sort.Slice(valSlice, func(i, j int) bool {
			return valSlice[i].AnnualRate.GT(valSlice[j].AnnualRate)
		})
		return valSlice, nil
	}

	shouldRm := (initialLen - int(number)) / 2
	valSlice = valSlice[shouldRm : initialLen-shouldRm]

	sort.Slice(valSlice, func(i, j int) bool {
		return valSlice[i].AnnualRate.GT(valSlice[j].AnnualRate)
	})
	if len(valSlice) > int(number) {
		valSlice = valSlice[:len(valSlice)-1]
	}

	return valSlice, nil
}

func GetValidatorAnnualRate(c *cosmosClient.Client, height int64) (map[string]*Validator, error) {
	blockResults, err := c.GetBlockResults(height)
	if err != nil {
		return nil, err
	}

	rewardMap := make(map[string]sdk.Dec, 0)
	for _, event := range sdk.StringifyEvents(blockResults.BeginBlockEvents) {
		if event.Type == "rewards" {
			if len(event.Attributes)%2 != 0 {
				return nil, fmt.Errorf("atribute len err, event: %s", event)
			}

			for i := 0; i < len(event.Attributes); i += 2 {
				rewardToken, err := sdk.ParseDecCoin(event.Attributes[i].Value)
				if err != nil {
					return nil, err
				}
				rewardMap[event.Attributes[i+1].Value] = rewardToken.Amount
			}
		}
	}

	res, err := c.QueryValidators(height)
	if err != nil {
		return nil, err
	}

	vals := make(map[string]*Validator, 0)
	for _, val := range res.Validators {
		willUseVal := Validator{
			Height:          height,
			OperatorAddress: val.OperatorAddress,
			TokenAmount:     val.Tokens,
			ShareAmount:     val.DelegatorShares,
			Commission:      val.GetCommission(),
		}

		if rewardTokenAmount, exist := rewardMap[val.OperatorAddress]; exist {

			commission := rewardTokenAmount.Mul(val.GetCommission())
			sharedToken := rewardTokenAmount.Sub(commission)

			rewardPerShare := sharedToken.Quo(willUseVal.ShareAmount)
			annualRate := rewardPerShare.Mul(sdk.NewDec(365 * 24 * 60 * 60)).Quo(averageBlockTIme)

			willUseVal.RewardAmount = rewardTokenAmount
			willUseVal.AnnualRate = annualRate

			vals[willUseVal.OperatorAddress] = &willUseVal
		}
	}

	return vals, nil
}

type Validator struct {
	Height          int64
	OperatorAddress string
	TokenAmount     sdk.Int
	RewardAmount    sdk.Dec
	ShareAmount     sdk.Dec
	Commission      sdk.Dec
	AnnualRate      sdk.Dec
}

type WrapMap struct {
	Cache      map[string]string
	CacheMutex sync.RWMutex
}
