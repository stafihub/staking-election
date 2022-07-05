package election_handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stafihub/staking-election/dao/election"
	"github.com/stafihub/staking-election/utils"
)

type RspAnnualRateList struct {
	AnnualRateList []AnnualRate `json:"annualRateList"`
}

type AnnualRate struct {
	RTokenDenom string  `json:"rTokenDenom"`
	AnnualRate  float64 `json:"annualRate"`
}

// @Summary get rate info
// @Description get annual rate info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspAnnualRateList}
// @Router /v1/annualRateList [get]
func (h *Handler) HandleGetAverageAnnualRate(c *gin.Context) {

	annualRateList, err := dao_election.GetAnnualRateList(h.db)
	if err != nil {
		logrus.Error("dao_election.GetAnnualRateList err: %s", err)
		utils.Err(c, codeInternalErr, err.Error())
		return
	}

	rsp := RspAnnualRateList{
		AnnualRateList: make([]AnnualRate, len(annualRateList)),
	}
	for i, rate := range annualRateList {
		dec, err := decimal.NewFromString(rate.AnnualRate)
		if err != nil {
			logrus.Error("dao_election.GetAnnualRateList err: %s", err)
			utils.Err(c, codeInternalErr, err.Error())
			return
		}

		rsp.AnnualRateList[i] = AnnualRate{
			RTokenDenom: rate.RTokenDenom,
			AnnualRate:  dec.InexactFloat64(),
		}
	}

	utils.Ok(c, "success", rsp)
}
