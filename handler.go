package main

import (
	"strings"
	"os"
	"text/template"
	"net/http"
	"path/filepath"
)

type Handler struct {
	Cache string
	Html string
	Repo string
	Sh string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.HandleIndex(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/sh/") {
		h.HandleSh(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/repo/") {
		h.HandleRepo(w, r)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/") {
		h.HandleDir(w, r)
		return
	} else {
		h.HandleCache(w, r)
	}

}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "html")
	http.ServeFile(w, r, h.Html + "index.html")
}

func (h *Handler) HandleRepo(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "repo")
	p := h.Repo + strings.TrimPrefix(r.URL.Path, "/repo/")
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
	}
}

func (h *Handler) HandleSh(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "sh")
	p := h.Sh + strings.TrimPrefix(r.URL.Path, "/sh/")
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
	}
}

func (h *Handler) HandleDir(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "dir")
	handleOrigin(w, r, "")
}

func (h *Handler) HandleCache(w http.ResponseWriter, r *http.Request) {
	p := h.Cache + strings.TrimLeft(r.URL.Path, "/")
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		key, err := filepath.Abs(p)
		if err != nil {
			if locked := filelock.RLock(key); locked {
				defer filelock.RUnlock(key)
				logRequest(r, "read cache")
				http.ServeFile(w, r, p)
				return
			} else {
				logRequest(r, "read origin")
				handleOrigin(w, r, "")
				return
			}
		}
	}
	logRequest(r, "read origin and write cache")
	handleOrigin(w, r, p)
}