package server

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stafihub/staking-election/api"
	"github.com/stafihub/staking-election/config"
	"github.com/stafihub/staking-election/utils"
)

type Server struct {
	listenAddr string
	httpServer *http.Server
	cfg        *config.Config
}

func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		listenAddr: cfg.ListenAddr,
		cfg:        cfg,
	}

	cache := map[string]string{}

	handler := s.InitHandler(cache)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Server) InitHandler(cache map[string]string) http.Handler {
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
	utils.SafeGoWithRestart(svr.ApiServer)
	return nil
}

func (svr *Server) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
}
