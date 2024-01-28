package main

import (
	"os"

	"github.com/freshdesk/resque_exporter"
)

func main() {
	resqueExporter.Run(os.Args[1:])
}
