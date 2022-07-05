package election_handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stafihub/staking-election/dao/election"
	"github.com/stafihub/staking-election/utils"
)

type RspSelectedValidators struct {
	SelectedValidators []SelectedValidator `json:"selectedValidators"`
}

type SelectedValidator struct {
	RTokenDenom      string `json:"rTokenDenom"`
	ValidatorAddress string `json:"validatorAddress"`
	Moniker          string `json:"moniker"`
	LogoUrl          string `json:"logo_url"`
}

// @Summary get selected validators
// @Description get selected validators
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspSelectedValidators}
// @Router /v1/selectedValidators [get]
func (h *Handler) HandleGetSelectedValidators(c *gin.Context) {

	selectedValidators, err := dao_election.GetAllSelectedValidators(h.db)
	if err != nil {
		logrus.Error("dao_election.GetAllSelectedValidators err: %s", err)
		utils.Err(c, codeInternalErr, err.Error())
		return
	}

	rsp := RspSelectedValidators{
		SelectedValidators: make([]SelectedValidator, len(selectedValidators)),
	}
	for i, val := range selectedValidators {
		rsp.SelectedValidators[i] = SelectedValidator{
			RTokenDenom:      val.RTokenDenom,
			ValidatorAddress: val.ValidatorAddress,
			Moniker:          val.Moniker,
			LogoUrl:          "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/moniker/cosmoshub/" + val.ValidatorAddress + ".png",
		}
	}

	utils.Ok(c, "success", rsp)
}
