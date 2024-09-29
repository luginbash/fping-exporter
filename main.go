package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var opts struct {
	Listen  string `short:"l" long:"listen" description:"Listen address" value-name:"[HOST]:PORT" default:":9605"`
	Period  uint   `short:"p" long:"period" description:"Period in seconds, should match Prometheus scrape interval" value-name:"SECS" default:"60"`
	Fping   string `short:"f" long:"fping"  description:"Fping binary path, leave blank to lookup in the PATH" value-name:"PATH"`
	Count   uint   `short:"c" long:"count"  description:"Number of pings to send at each period" value-name:"N" default:"20"`
	Version bool   `long:"version" description:"Show version"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func probeHandler(w http.ResponseWriter, r *http.Request) {
	targetParam := r.URL.Query().Get("target")
	if targetParam == "" {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html>
		    <head><title>Fping Exporter</title></head>
			<body>
			<b>ERROR: missing target parameter</b>
			</body>`))
		return
	}

	target := GetTarget(
		WorkerSpec{
			period: time.Second * time.Duration(opts.Period),
		},
		TargetSpec{
			host: targetParam,
		},
	)

	h := promhttp.HandlerFor(target.registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}
	if opts.Version {
		fmt.Printf("fping-exporter %v (commit %v, built %v)\n", buildVersion, buildCommit, buildDate)
		os.Exit(0)
	}

	if opts.Fping == "" {
		fpingPath, err := exec.LookPath("fping")
		if err != nil {
			log.Fatal("error looking up for fping")
		}
		opts.Fping = fpingPath
		log.Printf("using fping from PATH: %s\n", opts.Fping)
	}

	if _, err := os.Stat(opts.Fping); os.IsNotExist(err) {
		fmt.Printf("could not find fping at %q\n", opts.Fping)
		os.Exit(1)
	}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
