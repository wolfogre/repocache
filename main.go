package main

import (
	"flag"
	"net/http"
	"log"
)

var (
	bind = flag.String("bind", "0.0.0.0:80", "Bind address")
)

func main() {
	flag.Parse()
	patrol.Start("cache/")
	defer patrol.Stop()
	log.Println(http.ListenAndServe(*bind, &Handler{
		Cache: "cache/",
		Html: "html/",
		Repo: "repo/",
		Sh: "sh/",
	}))
}


