package server

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	cosmosClient "github.com/stafihub/cosmos-relay-sdk/client"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/stafihub/staking-election/api"
	"github.com/stafihub/staking-election/config"
	"github.com/stafihub/staking-election/utils"
)

type Server struct {
	listenAddr      string
	httpServer      *http.Server
	stop            chan struct{}
	cfg             *config.Config
	cache           *utils.WrapMap
	cosmosClientMap map[string]*cosmosClient.Client
	stafihubClient  *stafihubClient.Client
}

func NewServer(cfg *config.Config, stafihubClient *stafihubClient.Client) (*Server, error) {
	s := &Server{
		listenAddr: cfg.ListenAddr,
		cfg:        cfg,
		stop:       make(chan struct{}),
		cache: &utils.WrapMap{
			Cache: make(map[string]string),
		},
		stafihubClient: stafihubClient,
	}

	handler := s.InitHandler(s.cache)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Server) InitHandler(cache *utils.WrapMap) http.Handler {
	return api.InitRouters(cache)
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
	}
	for denom, client := range svr.cosmosClientMap {
		height, err := client.GetCurrentBlockHeight()
		if err != nil {
			return err
		}
		rate, err := utils.GetAverageAnnualRate(client, height, nil)
		if err != nil {
			return err
		}

		svr.cache.CacheMutex.Lock()
		svr.cache.Cache[denom] = rate.String()
		svr.cache.CacheMutex.Unlock()
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
			for denom, client := range s.cosmosClientMap {
				height, err := client.GetCurrentBlockHeight()
				if err != nil {
					continue
				}
				rate, err := utils.GetAverageAnnualRate(client, height, nil)
				if err != nil {
					continue
				}

				s.cache.CacheMutex.Lock()
				s.cache.Cache[denom] = rate.String()
				s.cache.CacheMutex.Unlock()
				logrus.Debugf("got average rate: %s", rate.String())
			}
			logrus.Debugf("AverageAnnualRateHandler end -----------")
		}
	}
}
