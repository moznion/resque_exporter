package resqueExporter

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mkideal/cli"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	rev string
	ver string
)

type opt struct {
	Help    bool   `cli:"h,help" usage:"display help"`
	Version bool   `cli:"v,version" usage:"display version and revision"`
	Port    int    `cli:"p,port" usage:"set port number" dft:"5555"`
	Config  string `cli:"c,config" usage:"set path to config file"`
}

func Run(args []string) {
	var configPath string
	var port int

	cli.Run(&opt{}, func(ctx *cli.Context) error {
		argv := ctx.Argv().(*opt)
		if argv.Help {
			ctx.String(ctx.Usage())
			os.Exit(0)
		}

		if argv.Version {
			ctx.String(fmt.Sprintf("%s/%s\n", ver, rev))
			os.Exit(0)
		}

		if argv.Config == "" {
			ctx.String(ctx.Usage())
		}

		configPath = argv.Config
		port = argv.Port

		return nil
	})

	if configPath == "" {
		log.Fatal("Missing mandatory option parameter: --config")
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	exporter, err := newExporter(config)
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("%s/%s\n", ver, rev)))
	})

	addr := fmt.Sprintf(":%d", port)
	log.Print("Listening 0.0.0.0", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
