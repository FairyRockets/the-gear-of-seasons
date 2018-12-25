package web

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fairy-rockets/the-gear-of-seasons/shelf"
	"github.com/julienschmidt/httprouter"
)

func (srv *Server) serveMoment(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.URL.Path == "/moment/search" {
		srv.serveMomentSearch(w, r, p)
		return
	}
	m := srv.shelf.LookupMoment(strings.TrimPrefix(r.URL.Path, "/moment"))
	if m == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	body := srv.momentCache.Fetch(m).Content()
	w.Write([]byte(body))
}

type momentSummary struct {
	Angle    float64 `json:"angle"`
	Date     string  `json:"date"`
	Title    string  `json:"title"`
	Path     string  `json:"path"`
	ImageURL string  `json:"imageURL"`
	BodyURL  string  `json:"bodyURL"`
}

func (srv *Server) makeSummary(m *shelf.Moment) *momentSummary {
	var err error
	var img *shelf.ImageEntity
	img = srv.momentCache.Fetch(m).FindFirstImage()
	beg := time.Date(m.Date.Year(), time.January, 1, 0, 0, 0, 0, m.Date.Location())
	end := time.Date(m.Date.Year()+1, time.January, 1, 0, 0, 0, 0, m.Date.Location())
	angle := float64(m.Date.Sub(beg)) / float64(end.Sub(beg))

	imageURL := ""
	if img != nil {
		_, err = srv.entityCache.FetchIcon(img)
		if err == nil {
			imageURL = fmt.Sprintf("/entity/%s/icon", img.ID())
		}
	}

	return &momentSummary{
		Angle:    angle * math.Pi * 2,
		Date:     strings.Replace(m.DateString(), "\n", "<br>", -1),
		Title:    m.Title,
		Path:     m.Path(),
		ImageURL: imageURL,
		BodyURL:  fmt.Sprintf("/moment%s", m.Path()),
	}
}

func (srv *Server) serveMomentSearch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		size = 100
		return
	}
	pi2 := math.Pi * 2.0
	lst := make([][]*momentSummary, int(math.Ceil(math.Sqrt(float64(size)))))
	orig := srv.shelf.FindAllMoments()
	if size > len(orig) {
		size = len(orig)
	}
	for _, v := range rand.Perm(len(orig)) {
		s := srv.makeSummary(orig[v])
		i := int(math.Round(s.Angle*float64(len(lst))/pi2)) % len(lst)
		lst[i] = append(lst[i], s)
	}
	out := make([]*momentSummary, size)
	cnt := 0
	step := 0
end:
	for cnt < size {
		for _, v := range lst {
			if step < len(v) {
				out[cnt] = v[step]
				cnt++
				if cnt >= size {
					break end
				}
			}
		}
		step++
	}
	body, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		srv.setError(w, r, err)
		return
	}
	w.Write(body)
}
