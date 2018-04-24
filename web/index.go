package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (srv *Web) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	t, err := srv.templateOf("index.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		srv.setError(w, r, err)
	}
}
