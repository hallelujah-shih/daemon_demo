package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"syscall"
)

func (s *ApiService) handlerSignal() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rspData []byte
		var err error
		var sig int64
		processName := strings.TrimSpace(r.URL.Query().Get("name"))
		if processName == "" {
			rspData, err = rspErrData(errors.New("lost process name"))
		}
		sigStr := strings.TrimSpace(r.URL.Query().Get("sig"))
		// sigStr == ""默认为0，当作探测进程是否running
		if sigStr != "" {
			// 此处没有对sig_num做任何校验
			sig, err = strconv.ParseInt(sigStr, 10, 8)
		}
		if err != nil {
			rspData, err = rspErrData(err)
		} else {
			s.pm.FwdSignal(processName, syscall.Signal(sig))
			// 此处不必将signal的异常报告client
			rspData, err = rspOkData()
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(rspData)
	})
}
