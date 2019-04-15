package api

import (
	"daemon_demo/config"
	"daemon_demo/process"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type ResponseData struct {
	Status int         `json:"status"`
	Error  string      `json:"error"`
	Data   interface{} `json:"data"`
}

func rspErrData(err error) ([]byte, error) {
	var rsp = &ResponseData{}
	rsp.Status = 500
	rsp.Error = err.Error()
	return json.Marshal(rsp)
}

func rspOkData() ([]byte, error) {
	var rsp = &ResponseData{}
	rsp.Status = 200
	return json.Marshal(rsp)
}

type ApiService struct {
	cfg    *config.ManagerConfig
	pm     *process.Manager
	router *mux.Router
}

func (s *ApiService) Serve() *http.Server {
	srv := &http.Server{
		Handler:      s.router,
		Addr:         s.cfg.ServeAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Minute,
	}
	return srv
}

func New(cfg *config.ManagerConfig, pm *process.Manager) *ApiService {
	apiSvc := &ApiService{
		cfg:    cfg,
		pm:     pm,
		router: mux.NewRouter(),
	}
	apiSvc.router.Handle("/process", apiSvc.handlerList())
	apiSvc.router.Handle("/process/stop", apiSvc.handlerStop())
	apiSvc.router.Handle("/process/start", apiSvc.handlerStart())
	apiSvc.router.Handle("/process/signal", apiSvc.handlerSignal())
	return apiSvc
}
