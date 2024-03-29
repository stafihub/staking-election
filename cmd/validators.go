package cmd

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	client "github.com/stafihub/cosmos-relay-sdk/client"
	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/staking-election/utils"
)

const flagNode = "node"
const flagNumber = "number"
const flagMaxMissedBlocks = "max-missed-blocks"

func selectValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "select-vals",
		Aliases: []string{"v"},
		Args:    cobra.ExactArgs(0),
		Short:   "Select high quality validators for you",
		RunE: func(cmd *cobra.Command, args []string) error {

			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}
			logrus.SetLevel(logLevel)
			node, err := cmd.Flags().GetString(flagNode)
			if err != nil {
				return err
			}

			fmt.Println("node rpc: ", node)
			prefix, err := cmd.Flags().GetString(flagPrefix)
			if err != nil {
				return err
			}
			fmt.Println("prefix: ", prefix)
			number, err := cmd.Flags().GetInt64(flagNumber)
			if err != nil {
				return err
			}
			fmt.Println("number: ", number)
			maxMissedBlocks, err := cmd.Flags().GetInt64(flagMaxMissedBlocks)
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
			fmt.Println("currentBlockHeight: ", curBLockHeight)

			fmt.Println("wait to get allValidator...")
			allValidator, err := utils.GetValidatorAnnualRate(c, curBLockHeight)
			if err != nil {
				return err
			}

			fmt.Println("wait to get averageAnnualRate...")
			averageAnnualRate, err := utils.GetAverageAnnualRate(c, curBLockHeight, allValidator)
			if err != nil {
				return err
			}
			fmt.Println("wait to get selectedValidator...")
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
				// (0). should remove slashed validators
				if slashRes.Pagination.Total > utils.MaxSlashAmount {
					fmt.Printf("slash total: %d, will skip \n", slashRes.Pagination.Total)
					continue
				}

				// (1). should skip missed blocks excessively validator
				validatorRes, err := c.QueryValidator(val.OperatorAddress, curBLockHeight)
				if err != nil {
					return err
				}

				done := core.UseSdkConfigContext(c.GetAccountPrefix())
				consPubkeyJson, err := c.Ctx().Codec.MarshalJSON(validatorRes.Validator.ConsensusPubkey)
				if err != nil {
					done()
					return err
				}
				var pk cryptotypes.PubKey
				if err := c.Ctx().Codec.UnmarshalInterfaceJSON(consPubkeyJson, &pk); err != nil {
					done()
					return err
				}
				consAddr := sdk.ConsAddress(pk.Address())
				consAddrStr := consAddr.String()
				done()

				signInfo, err := c.QuerySigningInfo(consAddrStr, curBLockHeight)
				if err != nil {
					return err
				}
				if validatorRes.Validator.Jailed || signInfo.ValSigningInfo.Tombstoned || signInfo.ValSigningInfo.MissedBlocksCounter > maxMissedBlocks {
					fmt.Printf("validator jailed %v, tombstoned %v, missed %d, will skip \n", validatorRes.Validator.Jailed, signInfo.ValSigningInfo.Tombstoned, signInfo.ValSigningInfo.MissedBlocksCounter)
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
	cmd.Flags().Int64(flagMaxMissedBlocks, 100, "max missed blocks")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")

	return cmd
}

func showValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-val [val-address]",
		Aliases: []string{"v"},
		Args:    cobra.ExactArgs(1),
		Short:   "Show validator info",
		RunE: func(cmd *cobra.Command, args []string) error {

			node, err := cmd.Flags().GetString(flagNode)
			if err != nil {
				return err
			}
			fmt.Println("node rpc: ", node)
			prefix, err := cmd.Flags().GetString(flagPrefix)
			if err != nil {
				return err
			}
			fmt.Println("prefix: ", prefix)

			valAddr := args[0]

			c, err := client.NewClient(nil, "", "", prefix, []string{node})
			if err != nil {
				return err
			}

			curBLockHeight, err := c.GetCurrentBlockHeight()
			if err != nil {
				return err
			}
			fmt.Println("currentBlockHeight: ", curBLockHeight)

			fmt.Println("wait to get allValidator...")
			allValidator, err := utils.GetValidatorAnnualRate(c, curBLockHeight)
			if err != nil {
				return err
			}
			fmt.Println("wait to get averageAnnualRate...")
			averageAnnualRate, err := utils.GetAverageAnnualRate(c, curBLockHeight, allValidator)
			if err != nil {
				return err
			}

			val := allValidator[valAddr]
			fmt.Println("total validators: ", len(allValidator))
			fmt.Println("average annual rate: ", averageAnnualRate.String())
			fmt.Printf("valAddress: %s annualRate: %s commission: %s tokenAmount: %s\n", val.OperatorAddress, val.AnnualRate, val.Commission, val.TokenAmount)

			return nil
		},
	}

	cmd.Flags().String(flagNode, "http://localhost:26657", "Node rpc endpoint")
	cmd.Flags().String(flagPrefix, "cosmos", "Account prefix (comos|stafi|iaa)")

	return cmd
}
