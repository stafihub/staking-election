package cmd

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	client "github.com/stafihub/cosmos-relay-sdk/client"
	"github.com/stafihub/staking-election/utils"
)

const flagNode = "node"
const flagNumber = "number"

func validatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validators",
		Aliases: []string{"v"},
		Short:   "Select high quality validators for you",
		RunE: func(cmd *cobra.Command, args []string) error {

			node, err := cmd.Flags().GetString(flagNode)
			if err != nil {
				return err
			}
			prefix, err := cmd.Flags().GetString(flagPrefix)
			if err != nil {
				return err
			}
			number, err := cmd.Flags().GetInt64(flagNumber)
			if err != nil {
				return err
			}

			c, err := client.NewClient(nil, "", "", prefix, []string{node})
			if err != nil {
				return err
			}

			curBLockHeight, err := c.GetCurrentBlockHeight()
			if err != nil {
				return err
			}

			allValidator, err := utils.GetValidatorAnnualRate(c, curBLockHeight)
			if err != nil {
				return err
			}

			averageAnnualRate, err := utils.GetAverageAnnualRate(c, curBLockHeight, allValidator)
			if err != nil {
				return err
			}
			valSlice, err := utils.GetSelectedValidator(c, curBLockHeight, number, allValidator)
			if err != nil {
				return err
			}
			fmt.Println("total validators: ", len(allValidator))
			fmt.Println("average annual rate: ", averageAnnualRate.String())
			fmt.Println("\nselected validators: ")
			for _, val := range valSlice {
				valAddress, err := sdk.ValAddressFromBech32(val.OperatorAddress)
				if err != nil {
					return err
				}
				slashRes, err := c.QueryValidatorSlashes(valAddress, curBLockHeight-utils.SlashDuBlock, curBLockHeight)
				if err != nil {
					return err
				}
				//should remove slashed validators
				if slashRes.Pagination.Total > utils.MaxSlashAmount {
					continue
				}

				fmt.Printf("valAddress: %s annualRate: %s commission: %s tokenAmount: %s\n", val.OperatorAddress, val.AnnualRate, val.Commission, val.TokenAmount)
			}

			return nil
		},
	}

	cmd.Flags().String(flagNode, "http://localhost:26657", "Node rpc endpoint")
	cmd.Flags().Int64(flagNumber, 5, "Validators number limit")
	cmd.Flags().String(flagPrefix, "cosmos", "Account prefix (comos|stafi|iaa)")

	return cmd
}
