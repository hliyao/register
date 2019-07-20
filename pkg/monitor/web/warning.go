package web

import (
	"encoding/json"
	"golang.52tt.com/pkg/monitor"
	"net/http"
)

func WarningHandle(w http.ResponseWriter, r *http.Request) {
	as := monitor.Status()
	data, err := json.Marshal(as)
	if err != nil {
		w.Write([]byte{})
		return
	}
	w.Write(data)
}