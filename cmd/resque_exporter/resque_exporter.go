package main

import (
	"os"

	"github.com/moznion/resque_exporter"
)

func main() {
	resqueExporter.Run(os.Args[1:])
}
