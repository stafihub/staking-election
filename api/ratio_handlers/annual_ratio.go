package ratio_handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stafihub/staking-election/utils"
)

type RspAnnualRatioList struct {
	AnnualRatioList []AnnualRatio `json:"annualRatioList"`
}

type AnnualRatio struct {
	RTokenDenom string `json:"rTokenDenom"`
	AnnualRatio string `json:"annualRatio"`
}

// @Summary get pool info
// @Description get pool info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspAnnualRatioList}
// @Router /v1/annualRatioList [get]
func (h *Handler) HandleGetAverageAnnualRatio(c *gin.Context) {
	rsp := RspAnnualRatioList{
		AnnualRatioList: []AnnualRatio{{"uratom", "0.16"}},
	}
	utils.Ok(c, "success", rsp)
}
