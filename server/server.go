package server

import (
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
	"github.com/stafihub/rtoken-relay-core/common/core"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/stafihub/staking-election/api"
	"github.com/stafihub/staking-election/config"
	"github.com/stafihub/staking-election/dao/election"
	"github.com/stafihub/staking-election/db"
	"github.com/stafihub/staking-election/utils"
	"gorm.io/gorm"
)

type Server struct {
	listenAddr      string
	httpServer      *http.Server
	stop            chan struct{}
	cfg             *config.Config
	db              *db.WrapDb
	cosmosClientMap map[string]*cosmosClient.Client
	stafihubClient  *stafihubClient.Client
}

func NewServer(cfg *config.Config, stafihubClient *stafihubClient.Client, db *db.WrapDb) (*Server, error) {
	s := &Server{
		listenAddr:     cfg.ListenAddr,
		cfg:            cfg,
		stop:           make(chan struct{}),
		stafihubClient: stafihubClient,
		db:             db,
	}

	handler := s.InitHandler(s.db)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Server) InitHandler(db *db.WrapDb) http.Handler {
	return api.InitRouters(db)
}

func (svr *Server) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.listenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.listenAddr)
}

func (svr *Server) Start() error {
	svr.cosmosClientMap = make(map[string]*cosmosClient.Client)
	// init client and selected validators
	for _, rtokenInfo := range svr.cfg.RTokenInfo {
		addressPrefixRes, err := svr.stafihubClient.QueryAddressPrefix(rtokenInfo.Denom)
		if err != nil {
			return err
		}
		client, err := cosmosClient.NewClient(nil, "", "", addressPrefixRes.AccAddressPrefix, rtokenInfo.EndpointList)
		if err != nil {
			return err
		}
		svr.cosmosClientMap[rtokenInfo.Denom] = client

		bondedPoolsRes, err := svr.stafihubClient.QueryPools(rtokenInfo.Denom)
		if err != nil {
			return err
		}

		for _, poolAddrStr := range bondedPoolsRes.Addrs {
			done := core.UseSdkConfigContext(client.GetAccountPrefix())
			poolAddr, err := sdk.AccAddressFromBech32(poolAddrStr)
			if err != nil {
				done()
				return err
			}
			done()

			delegationsRes, err := client.QueryDelegations(poolAddr, 0)
			if err != nil {
				return err
			}

			for _, delegation := range delegationsRes.DelegationResponses {
				valAddress := delegation.Delegation.ValidatorAddress
				_, err := dao_election.GetSelectedValidator(svr.db, rtokenInfo.Denom, poolAddrStr, valAddress)
				if err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					} else {
						valRes, err := client.QueryValidator(valAddress, 0)
						if err != nil {
							return err
						}

						selectedValidator := &dao_election.SelectedValidator{
							RTokenDenom:      rtokenInfo.Denom,
							PoolAddress:      poolAddrStr,
							ValidatorAddress: valAddress,
							Moniker:          valRes.Validator.GetMoniker(),
						}

						err = dao_election.UpOrInSelectedValidator(svr.db, selectedValidator)
						if err != nil {
							return err
						}
					}
				}

			}
		}
	}

	// init annual rate
	for denom, client := range svr.cosmosClientMap {
		height, err := client.GetCurrentBlockHeight()
		if err != nil {
			return err
		}
		rate, err := utils.GetAverageAnnualRate(client, height, nil)
		if err != nil {
			return err
		}
		annualRate, err := dao_election.GetAnnualRate(svr.db, denom)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		annualRate.RTokenDenom = denom
		annualRate.AnnualRate = rate.String()

		err = dao_election.UpOrInAnnualRate(svr.db, annualRate)
		if err != nil {
			return err
		}
	}

	utils.SafeGoWithRestart(svr.ApiServer)
	utils.SafeGoWithRestart(svr.AverageAnnualRateHandler)
	return nil
}

func (svr *Server) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
	close(svr.stop)
}

func (s *Server) AverageAnnualRateHandler() {
	logrus.Infof("AverageAnnualRateHandler start")
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			logrus.Debugf("AverageAnnualRateHandler start -----------")
			err := s.updateAnnualRate()
			if err != nil {
				logrus.Warnf("updateAnnualRate err: %s", err)
				continue
			}
			logrus.Debugf("AverageAnnualRateHandler end -----------")

			logrus.Debugf("updateSelectedValidator start -----------")
			err = s.updateSelectedValidator()
			if err != nil {
				logrus.Warnf("updateSelectedValidator err: %s", err)
				continue
			}
			logrus.Debugf("updateSelectedValidator end -----------")
		}
	}
}

func (svr *Server) updateAnnualRate() error {
	for denom, client := range svr.cosmosClientMap {
		height, err := client.GetCurrentBlockHeight()
		if err != nil {
			return err
		}
		rate, err := utils.GetAverageAnnualRate(client, height, nil)
		if err != nil {
			return err
		}

		annualRate, err := dao_election.GetAnnualRate(svr.db, denom)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		annualRate.RTokenDenom = denom
		annualRate.AnnualRate = rate.String()

		err = dao_election.UpOrInAnnualRate(svr.db, annualRate)
		if err != nil {
			return err
		}

		logrus.Debugf("got average rate: %s", rate.String())
	}
	return nil
}

func (svr *Server) updateSelectedValidator() error {
	for _, rtokenInfo := range svr.cfg.RTokenInfo {

		client := svr.cosmosClientMap[rtokenInfo.Denom]

		bondedPoolsRes, err := svr.stafihubClient.QueryPools(rtokenInfo.Denom)
		if err != nil {
			return err
		}

		for _, poolAddrStr := range bondedPoolsRes.Addrs {
			done := core.UseSdkConfigContext(client.GetAccountPrefix())
			poolAddr, err := sdk.AccAddressFromBech32(poolAddrStr)
			if err != nil {
				done()
				return err
			}
			done()

			delegationsRes, err := client.QueryDelegations(poolAddr, 0)
			if err != nil {
				return err
			}

			for _, delegation := range delegationsRes.DelegationResponses {
				valAddress := delegation.Delegation.ValidatorAddress
				_, err := dao_election.GetSelectedValidator(svr.db, rtokenInfo.Denom, poolAddrStr, valAddress)
				if err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					} else {
						valRes, err := client.QueryValidator(valAddress, 0)
						if err != nil {
							return err
						}

						selectedValidator := &dao_election.SelectedValidator{
							RTokenDenom:      rtokenInfo.Denom,
							PoolAddress:      poolAddrStr,
							ValidatorAddress: valAddress,
							Moniker:          valRes.Validator.GetMoniker(),
						}

						err = dao_election.UpOrInSelectedValidator(svr.db, selectedValidator)
						if err != nil {
							return err
						}
					}
				}

			}
		}
	}
	return nil
}
