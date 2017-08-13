package main

import (
	"time"
	"path/filepath"
	"os"
	"errors"
	"net/http"
	"strings"
	"log"
)

var patrol = &Patrol{
	started: false,
	stopped: false,
}

type Patrol struct {
	started bool
	stopped bool
	path string
}

func (p *Patrol) Start(path string) {
	if !p.started {
		p.path = path
		p.started = true
		go p.worker()
	}
}

func (p *Patrol) Stop() {
	p.stopped = true
}

func (p *Patrol) worker() {
	for !p.stopped {
		err := filepath.Walk(p.path, func (path string, f os.FileInfo, err error) error {
			path = strings.Replace(path, "\\", "/", -1)
			RETRY:
			if err != nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			time.Sleep(time.Second)
			if p.stopped {
				return errors.New("Stopped")
			}
			resp, err := http.Head("http://mirrors.aliyun.com/" + strings.TrimPrefix(path, p.path))
			if err != nil {
				println(err.Error())
				goto RETRY
			}
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
				println(resp.Status)
				goto RETRY
			}
			for !filelock.WLock(path) {
				time.Sleep(time.Second)
			}
			defer filelock.WUnlock(path)
			if resp.StatusCode == http.StatusNotFound {
				log.Printf("Remove %v because remoe don't exist\n", path)
				err := os.Remove(path)
				log.Println("Remove " + path)
				if err != nil {
					return err
				}
				return nil
			}
			ff, err := os.Stat(path)
			if err != nil {
				return err
			}
			t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", resp.Header.Get("Last-Modified"))
			if err != nil {
				return err
			}
			if !t.Equal(ff.ModTime()) {
				log.Printf("Remove %v because local time %v don't match remoe time %v \n", path, ff.ModTime().Local(), t.Local())
				err := os.Remove(path)
				log.Println("Remove " + path)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}

