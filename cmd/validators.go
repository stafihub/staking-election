package cmd

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	client "github.com/stafihub/cosmos-relay-sdk/client"
)

const flagNode = "node"

var averageBlockTIme = sdk.MustNewDecFromStr("6.77")

func validatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validators",
		Aliases: []string{"v"},
		Short:   "show validators",
		RunE: func(cmd *cobra.Command, args []string) error {

			node, err := cmd.Flags().GetString(flagNode)
			if err != nil {
				return err
			}

			c, err := client.NewClient(nil, "", "", "", []string{node})
			if err != nil {
				return err
			}

			curBLockHeight, err := c.GetCurrentBlockHeight()
			if err != nil {
				return err
			}

			valMap, err := GetValidatorAnnualRatio(c, curBLockHeight)
			if err != nil {
				return err
			}

			valSlice := make([]*Validator, 0)
			totalAnuualRatio := sdk.NewDec(0)
			for _, val := range valMap {
				valSlice = append(valSlice, val)
				totalAnuualRatio = totalAnuualRatio.Add(val.AnnualRatio)
			}

			sort.Slice(valSlice, func(i, j int) bool {
				return valSlice[i].TokenAmount.GT(valSlice[j].TokenAmount)
			})

			initialLen := len(valSlice)
			valSlice = valSlice[initialLen/10 : initialLen-initialLen/10]

			sort.Slice(valSlice, func(i, j int) bool {
				return valSlice[i].AnnualRatio.GT(valSlice[j].AnnualRatio)
			})
			valSlice = valSlice[:5]

			fmt.Println("total validators: ", initialLen)
			fmt.Println("average annual ratio: ", totalAnuualRatio.Quo(sdk.NewDec(int64(initialLen))))
			fmt.Println("selected validators: ")
			for _, val := range valSlice {
				fmt.Printf("%+v\n", val)
			}

			return nil
		},
	}

	cmd.Flags().String(flagNode, "", "node rpc endpoint")

	return cmd
}

func GetValidatorAnnualRatio(c *client.Client, height int64) (map[string]*Validator, error) {
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
			annualRatio := rewardPerShare.Mul(sdk.NewDec(365 * 24 * 60 * 60)).Quo(averageBlockTIme)

			willUseVal.RewardAmount = rewardTokenAmount
			willUseVal.AnnualRatio = annualRatio

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
	AnnualRatio     sdk.Dec
}
