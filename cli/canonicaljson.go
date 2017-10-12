package main

import (
	"flag"
	"fmt"
	"github.com/gibson042/canonicaljson-go"
	"log"
	"os"
)

func main() {
	flag.Parse()
	srcFiles := flag.Args()
	if len(srcFiles) == 0 {
		srcFiles = []string{"-"}
	}

	for _, srcFile := range srcFiles {
		var data interface{}
		var decoder *canonicaljson.Decoder

		if srcFile == "-" {
			decoder = canonicaljson.NewDecoder(os.Stdin)
		} else {
			file, err := os.Open(srcFile)
			if err != nil {
				log.Fatal(err)
			}
			decoder = canonicaljson.NewDecoder(file)
		}
		// Handle numbers with infinite precision.
		decoder.UseNumber()

		// Read as many JSON values as possible from standard input.
		for srcFile != "-" || decoder.More() {
			if err := decoder.Decode(&data); err != nil {
				log.Fatal(err)
			}

			if result, err := canonicaljson.Marshal(&data); err != nil {
				log.Fatal(err)
			} else {
				fmt.Printf("%s", string(result))
			}

			// Read only a single value from each file.
			if srcFile != "-" {
				if decoder.More() {
					log.Fatal("Trailing data in file: %s\n", srcFile)
				}
				break
			}
		}
	}
}
