package main

import (
	"time"
	"log"
)

func main() {
	source := "/mnt/ram/BAF_20180224.dat"
	outputdir := "/mnt/ram/posti"

	starttime := time.Now().UTC()
	ConvertFile(source, outputdir)
	endtime := time.Now().UTC()
	log.Printf("Took %s", endtime.Sub(starttime))
}
