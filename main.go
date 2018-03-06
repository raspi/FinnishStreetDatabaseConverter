package main

import (
	"time"
	"log"
	"flag"
	"os"
	"fmt"
)

func main() {
	var err error

	sourcefile := flag.String("f", "", "File name (BAF_yyyymmdd.dat)")
	outputdirectory := flag.String("o", "", "Output directory /home/user/jsonfiles")

	flag.Parse()

	required := []string{"f", "o"}

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})

	err = HasRequiredCommandLineArguments(required, seen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}

	_, err = os.Stat(*sourcefile)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "File not found: '%v'", *sourcefile)
			os.Exit(2)
		} else {
			panic(err)
		}

	}

	log.Printf("Source file: '%s'", *sourcefile)
	log.Printf("Output directory: '%s'", *outputdirectory)

	starttime := time.Now().UTC()

	err = ConvertFile(*sourcefile, *outputdirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: '%v'", err)
	}

	log.Printf("Took %s", time.Now().UTC().Sub(starttime))
}
