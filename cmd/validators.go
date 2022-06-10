package cmd

import (
	"fmt"
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
			number, err := cmd.Flags().GetInt64(flagNumber)
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

			allValidator, err := utils.GetValidatorAnnualRatio(c, curBLockHeight)
			if err != nil {
				return err
			}

			averageAnnualRatio, err := utils.GetAverageAnnualRatio(c, curBLockHeight)
			if err != nil {
				return err
			}
			valSlice, err := utils.GetSelectedValidator(c, curBLockHeight, number)
			if err != nil {
				return err
			}
			fmt.Println("total validators: ", len(allValidator))
			fmt.Println("average annual ratio: ", averageAnnualRatio.String())
			fmt.Println("\nselected validators: ")
			for _, val := range valSlice {
				fmt.Printf("%+v\n", val)
			}

			return nil
		},
	}

	cmd.Flags().String(flagNode, "", "node rpc endpoint")
	cmd.Flags().Int64(flagNumber, 5, "validators number limit")

	return cmd
}
