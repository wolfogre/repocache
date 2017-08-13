package main

import (
	"flag"
	"net/http"
	"log"
	"strings"
	"os"
	"path"
	"time"
	"text/template"
)

var (
	bind = flag.String("bind", "0.0.0.0:80", "Bind address")
)

func main() {
	flag.Parse()
	log.Println(http.ListenAndServe(*bind, &Handler{
		Cache: "./cache/",
		Html: "./html/",
		Repo: "./script/",
	}))
}

type Handler struct {
	Cache string
	Html string
	Repo string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "./html/index.html")
		log.Printf("%v %v [%v]", r.Method, r.URL.String(), "static")
		return
	}

	if strings.HasPrefix(r.URL.Path, "/repo/") {
		// TODO 有BUG
		p := h.Repo + strings.TrimPrefix(r.URL.Path, "/script/")
		t, err := template.ParseFiles(p)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		err = t.Execute(w, r.Host)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		return
	}

	if strings.HasSuffix(r.URL.Path, "/") {
		h.HandleOrigin(w, r, false)
		log.Printf("%v %v [%v]", r.Method, r.URL.String(), "dir")
		return
	} else {
		p := h.Cache + strings.TrimLeft(r.URL.Path, "/")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			if locked := filelock.RLock(p); locked {
				http.ServeFile(w, r, p)
				log.Printf("%v %v [%v]", r.Method, r.URL.String(), "read cache")
				return
			} else {
				h.HandleOrigin(w, r, false)
				log.Printf("%v %v [%v]", r.Method, r.URL.String(), "read origin")
				return
			}
		}
		h.HandleOrigin(w, r, true)
		log.Printf("%v %v [%v]", r.Method, r.URL.String(), "read origin and write cache")
	}

}

func (h *Handler) HandleOrigin(w http.ResponseWriter, r *http.Request, cache bool) {
	resp, err := requestOrigin(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	fillHeader(resp, w)
	if r.Method == "GET" {
		if !cache {
			fillBody(resp, w)
		} else {
			fillBodyAndCache(resp, w, h.Cache + strings.TrimLeft(r.URL.Path, "/"))
		}
	}
}

func (h *Handler) HandleScript(w http.ResponseWriter, r *http.Request, cache bool) {

}

func requestOrigin(r *http.Request) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, "http://mirrors.aliyun.com"  + r.URL.Path, r.Body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client.Do(req)
}

func fillHeader(rp *http.Response, w http.ResponseWriter) {
	for k, vs := range rp.Header {
		for _, v := range vs {
			if rp.StatusCode == http.StatusMovedPermanently && k == "Location" {
				w.Header().Add(k, strings.TrimPrefix(v, "http://mirrors.aliyun.com"))
				continue
			}
			w.Header().Add(k,v)
		}
	}
	w.WriteHeader(rp.StatusCode)
}

func fillBody(rp *http.Response, w http.ResponseWriter) {
	buffer := make([]byte, 4096, 4096)
	for n, _ := rp.Body.Read(buffer); n != 0; n, _ = rp.Body.Read(buffer) {
		_, err := w.Write(buffer[:n])
		if err != nil {
			break
		}
	}
	rp.Body.Close()
}

func fillBodyAndCache(rp *http.Response, w http.ResponseWriter, p string) {
	locked := filelock.WLock(p)
	if !locked {
		fillBody(rp, w)
		return
	}

	defer filelock.WUnlock(p)
	buffer := make([]byte, 4096, 4096)
	os.MkdirAll(path.Dir(p), os.ModeDir)
	f, err := os.Create(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	finished := true
	for n, _ := rp.Body.Read(buffer); n != 0; n, _ = rp.Body.Read(buffer) {
		_, err := w.Write(buffer[:n])
		if err != nil {
			finished = false
			break
		}
		_, err = f.Write(buffer[:n])
		if err != nil {
			finished = false
			break
		}
	}
	rp.Body.Close()
	f.Close()
	if !finished {
		os.Remove(p)
	} else {
		t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", rp.Header.Get("Last-Modified"))
		if err == nil {
			os.Chtimes(p, time.Now(), t)
		}
	}
}
