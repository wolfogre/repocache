package main

import (
	"path"
	"net/http"
	"strings"
	"log"
	"os"
	"time"
)

func handleOrigin(w http.ResponseWriter, r *http.Request, cache string) {
	resp, err := requestOrigin(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	fillHeader(resp, w)
	if r.Method == "GET" {
		if cache == "" {
			fillBody(resp, w)
		} else if resp.StatusCode == http.StatusOK {
			fillBodyAndCache(resp, w, cache)
		}
	}
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

func logRequest(r *http.Request, result string) {
	log.Printf("%v %v [%v]\n", r.Method, r.URL.String(), result)
}
