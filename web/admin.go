package web

import (
	"net/http"

	"encoding/json"
	"io/ioutil"

	"fmt"

	"strconv"

	"time"

	"strings"

	"sort"

	"github.com/fairy-rockets/the-gear-of-seasons/shelf"
	"github.com/fairy-rockets/the-gear-of-seasons/web/cache"
	"github.com/julienschmidt/httprouter"
)

func (srv *Server) serveAdminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	t, err := srv.templateOf("admin/_main.html", "admin/index.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		srv.setError(w, r, err)
	}

}

func (srv *Server) serveAdminNew(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	t, err := srv.templateOf("admin/_main.html", "admin/edit.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		srv.setError(w, r, err)
	}
}

func (srv *Server) serveAdminMoment(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	t, err := srv.templateOf("admin/_main.html", "admin/edit.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	m := srv.shelf.LookupMoment(strings.TrimPrefix(r.URL.Path, "/moment"))
	if m == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	err = t.Execute(w, m)
	if err != nil {
		srv.setError(w, r, err)
	}
}
func (srv *Server) serveAdminEdit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	t, err := srv.templateOf("admin/_main.html", "admin/edit.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		srv.setError(w, r, err)
	}
}

func (srv *Server) serveAdminPreview(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	m := new(shelf.Moment)
	err = json.Unmarshal(data, m)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	mc := srv.momentCache.Preview(m)
	w.WriteHeader(200)
	w.Write([]byte(mc.Content()))
}

func (srv *Server) serveAdminSave(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	m := new(shelf.Moment)
	err = json.Unmarshal(data, m)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	mc, err := srv.momentCache.Save(m)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	dat, err := json.Marshal(struct {
		Body string `json:"body"`
		Path string `json:"path"`
	}{
		Body: mc.Content(),
		Path: fmt.Sprintf("/moment%s", mc.Moment.Path()),
	})
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}

func (srv *Server) serveAdminUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	mimeType := r.Header.Get("Content-Type")
	if len(mimeType) == 0 {
		srv.setError(w, r, fmt.Errorf("empty Content-Type"))
		return
	}
	switch mimeType {
	case "image/jpeg":
		fallthrough
	case "image/png":
		fallthrough
	case "image/gif":
		/* Image */
		img, err := srv.shelf.AddImageEntity(mimeType, buffer)
		if err != nil {
			srv.setError(w, r, err)
			break
		}
		_, err = srv.entityCache.FetchMedium(img)
		if err != nil {
			srv.setError(w, r, err)
			break
		}
		_, err = srv.entityCache.FetchIcon(img)
		if err != nil {
			srv.setError(w, r, err)
			break
		}
		w.WriteHeader(200)
		fmt.Fprintf(w, "[image entity=\"%s\"]", img.ID)
	case "video/mp4":
		/* Video */
		w.WriteHeader(501)
		fmt.Fprintf(w, "Not Supported yet: %s\n", mimeType)
	default:
		w.WriteHeader(501)
		fmt.Fprintf(w, "Unknown Mime-Type: %s\n", mimeType)
	}
}

func (srv *Server) serveAdminMoments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, fmt.Sprintf("/moments/%d", time.Now().Year()), 302)
}
func (srv *Server) serveAdminMomentLists(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	t, err := srv.templateOf("admin/_main.html", "admin/moments.html")
	if err != nil {
		srv.setError(w, r, err)
		return
	}

	w.WriteHeader(200)
	year, err := strconv.Atoi(p.ByName("year"))
	if err != nil {
		year = time.Now().Year()
	}
	ms := srv.shelf.FindAllMomentsByYear(year)
	mcs := make([]*cache.MomentCache, 0, len(ms))
	for _, m := range ms {
		mcs = append(mcs, srv.momentCache.Fetch(m))
	}
	sort.Slice(mcs, func(i, j int) bool {
		return mcs[i].Moment.Date.After(mcs[j].Moment.Date)
	})
	err = t.Execute(w, struct {
		LastYear int
		Year     int
		NextYear int
		Moments  []*cache.MomentCache
	}{
		LastYear: year - 1,
		Year:     year,
		NextYear: year + 1,
		Moments:  mcs,
	})
	if err != nil {
		srv.setError(w, r, err)
	}
}
