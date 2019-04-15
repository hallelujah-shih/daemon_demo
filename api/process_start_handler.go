package api

import (
	"errors"
	"net/http"
	"strings"
)

func (s *ApiService) handlerStart() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rspData []byte
		var err error
		processName := strings.TrimSpace(r.URL.Query().Get("name"))
		if processName == "" {
			rspData, err = rspErrData(errors.New("lost process name"))
		} else {
			s.pm.ProcessStart(processName)
			rspData, err = rspOkData()
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(rspData)
	})
}
