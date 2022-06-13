package utils

import (
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
)

var (
	averageBlockTIme     = sdk.MustNewDecFromStr("6.77")
	stepNumber, stepSize = 8, 1000
	MaxSlashAmount       = uint64(0)
	SlashDuBlock         = int64(10000)
)

func GetAverageAnnualRate(c *cosmosClient.Client, height int64, valMap map[string]*Validator) (sdk.Dec, error) {
	var err error
	if valMap == nil {
		valMap, err = GetValidatorAnnualRate(c, height)
		if err != nil {
			return sdk.ZeroDec(), err
		}
	}

	totalAnuualRate := sdk.NewDec(0)
	for _, val := range valMap {
		totalAnuualRate = totalAnuualRate.Add(val.AnnualRate)
	}

	initialLen := len(valMap)

	return totalAnuualRate.Quo(sdk.NewDec(int64(initialLen))), nil
}

func GetSelectedValidator(c *cosmosClient.Client, height, number int64, valMap map[string]*Validator) ([]*Validator, error) {
	var err error
	if valMap == nil {
		valMap, err = GetValidatorAnnualRate(c, height)
		if err != nil {
			return nil, err
		}
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
	// rm 5% + 10%
	remainStart := initialLen / 20
	remainEnd := initialLen - initialLen/10
	if remainStart >= remainEnd || remainEnd-remainStart < int(number) {
		sort.Slice(valSlice, func(i, j int) bool {
			return valSlice[i].AnnualRate.GT(valSlice[j].AnnualRate)
		})
		return valSlice[:number], nil
	}
	valSlice = valSlice[remainStart:remainEnd]

	// selected by annualRate
	sort.SliceStable(valSlice, func(i, j int) bool {
		return valSlice[i].AnnualRate.GT(valSlice[j].AnnualRate)
	})
	valSlice = valSlice[:number]

	return valSlice, nil
}

func GetValidatorAnnualRate(c *cosmosClient.Client, height int64) (map[string]*Validator, error) {
	rates := make([]map[string]*Validator, 0)

	for i := 0; i < stepNumber; i++ {
		valRates, err := GetValidatorAnnualRateOnHeight(c, height-int64(i*stepSize))
		if err != nil {
			return nil, err
		}
		rates = append(rates, valRates)
	}

	retValRates := make(map[string]*Validator)
	for valAddr, val := range rates[0] {
		total := sdk.ZeroDec()
		for _, rate := range rates {
			if valRate, exist := rate[valAddr]; exist {
				total = total.Add(valRate.AnnualRate)
			}
		}
		val.AnnualRate = total.Quo(sdk.NewDec(int64(stepNumber)))
		retValRates[valAddr] = val
	}
	return retValRates, nil
}

func GetValidatorAnnualRateOnHeight(c *cosmosClient.Client, height int64) (map[string]*Validator, error) {
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
