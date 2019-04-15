package api

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ProcessList map[string]string

func (s *ApiService) handlerList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := s.pm.ProcessStatus()
		rsp := &ResponseData{}
		rsp.Error = ""
		rsp.Status = 200
		rsp.Data = status

		datas, err := json.Marshal(rsp)
		if err != nil {
			logrus.Errorln("rsp marshal error:", err)
			datas, err = rspErrData(err)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(datas)
	})
}
