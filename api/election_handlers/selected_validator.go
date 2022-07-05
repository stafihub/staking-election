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
	RTokenDenom   string            `json:"rTokenDenom"`
	ValidatorList []ValidatorDetail `json:"validatorList"`
}
type ValidatorDetail struct {
	ValidatorAddress string `json:"validatorAddress"`
	Moniker          string `json:"moniker"`
	LogoUrl          string `json:"logoUrl"`
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

	selectedValMap := make(map[string][]*dao_election.SelectedValidator)
	for _, val := range selectedValidators {
		selectedValMap[val.RTokenDenom] = append(selectedValMap[val.RTokenDenom], val)
	}

	rsp := RspSelectedValidators{
		SelectedValidators: make([]SelectedValidator, 0),
	}

	for key, vals := range selectedValMap {

		valList := make([]ValidatorDetail, 0)
		for _, val := range vals {
			valList = append(valList, ValidatorDetail{
				ValidatorAddress: val.ValidatorAddress,
				Moniker:          val.Moniker,
				LogoUrl:          "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/moniker/cosmoshub/" + val.ValidatorAddress + ".png",
			})
		}
		rsp.SelectedValidators = append(rsp.SelectedValidators, SelectedValidator{
			RTokenDenom:   key,
			ValidatorList: valList,
		})
	}

	utils.Ok(c, "success", rsp)
}
