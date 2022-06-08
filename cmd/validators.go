package cmd

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	client "github.com/stafihub/cosmos-relay-sdk/client"
)

const flagNode = "node"

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

			res, err := c.QueryValidators(0)
			if err != nil {
				return err
			}

			vals := make([]Validator, 0)
			for _, val := range res.Validators {
				vals = append(vals, Validator{
					OperatorAddress: val.OperatorAddress,
					TokenAmount:     val.Tokens,
					ShareAmount:     val.DelegatorShares,
					TokenPerShare:   val.Tokens.ToDec().Quo(val.DelegatorShares),
				})
			}

			sort.Slice(vals, func(i, j int) bool {
				return vals[i].TokenAmount.GT(vals[j].TokenAmount)
			})

			for _, val := range vals {
				fmt.Printf("%+v\n", val)
			}

			return nil
		},
	}

	cmd.Flags().String(flagNode, "", "node rpc endpoint")

	return cmd
}

type Validator struct {
	OperatorAddress string
	TokenAmount     sdk.Int
	ShareAmount     sdk.Dec
	TokenPerShare   sdk.Dec
}
