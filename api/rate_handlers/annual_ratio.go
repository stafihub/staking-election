package rate_handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stafihub/staking-election/utils"
)

type RspAnnualRateList struct {
	AnnualRateList []AnnualRate `json:"annualRateList"`
}

type AnnualRate struct {
	RTokenDenom string `json:"rTokenDenom"`
	AnnualRate  string `json:"annualRate"`
}

// @Summary get pool info
// @Description get pool info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspAnnualRateList}
// @Router /v1/annualRateList [get]
func (h *Handler) HandleGetAverageAnnualRate(c *gin.Context) {

	rsp := RspAnnualRateList{
		AnnualRateList: make([]AnnualRate, 0),
	}

	h.cache.CacheMutex.RLock()
	for denom, rate := range h.cache.Cache {
		rsp.AnnualRateList = append(rsp.AnnualRateList, AnnualRate{
			RTokenDenom: denom,
			AnnualRate:  rate,
		})
	}
	h.cache.CacheMutex.RUnlock()

	utils.Ok(c, "success", rsp)
}
