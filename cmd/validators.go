package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	client "github.com/stafihub/cosmos-relay-sdk/client"
)

func validatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validators",
		Aliases: []string{"v"},
		Short:   "show validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.NewClient(nil, "", "", "", []string{"https://cosmos-rpc1.stafi.io:443"})
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
					TokenAmount:     val.Tokens.String(),
				})
			}

			fmt.Println(vals)

			return nil
		},
	}

	return cmd
}

type Validator struct {
	OperatorAddress string
	TokenAmount     string
}
