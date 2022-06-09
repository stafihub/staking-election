// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package ratio_handlers

type Handler struct {
	cache map[string]string
}

func NewHandler(cache map[string]string) *Handler {
	return &Handler{cache: cache}
}

const (
	codeParamParseErr         = "80001"
	codeSymbolErr             = "80002"
	codeStafiAddressErr       = "80003"
	codeTxHashErr             = "80004"
	codePubkeyErr             = "80005"
	codeInternalErr           = "80006"
	codePoolAddressErr        = "80007"
	codeTxDuplicateErr        = "80008"
	codeTokenPriceErr         = "80009"
	codeInAmountFormatErr     = "80010"
	codeMinOutAmountFormatErr = "80011"
	codePriceSlideErr         = "80012"
	codeMinLimitErr           = "80013"
	codeMaxLimitErr           = "80014"
	codeSwapInfoNotExistErr   = "80015"
	codeLimitInfoNotExistErr  = "80016"
)
