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

	sourceFile := flag.String("f", "", "File name (BAF_yyyymmdd.dat)")
	outputDirectory := flag.String("o", "", "Output directory /home/user/jsonfiles")

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

	_, err = os.Stat(*sourceFile)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "File not found: '%v'", *sourceFile)
			os.Exit(2)
		} else {
			panic(err)
		}

	}

	log.Printf("Source file: '%s'", *sourceFile)
	log.Printf("Output directory: '%s'", *outputDirectory)

	starttime := time.Now().UTC()

	err = ConvertFile(*sourceFile, *outputDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: '%v'", err)
	}

	log.Printf("Took %s", time.Now().UTC().Sub(starttime))
}
