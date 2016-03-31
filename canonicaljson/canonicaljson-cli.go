package main

import (
	"encoding/json"
	"flag"
	"github.com/gibson042/canonicaljson-go"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	flag.Parse()
	srcFiles := flag.Args()
	if len(srcFiles) == 0 {
		srcFiles = []string{"-"}
	}

	encoder := canonicaljson.NewEncoder(os.Stdout)
	for _, srcFile := range srcFiles {
		var data interface{}
		var decoder *json.Decoder
		if srcFile == "-" {
			decoder = json.NewDecoder(os.Stdin)
		}
		for srcFile != "" && (srcFile != "-" || decoder.More()) {
			if srcFile == "-" {
				if err := decoder.Decode(&data); err != nil {
					log.Fatal(err)
				}
			} else {
				src, err := ioutil.ReadFile(srcFile)
				srcFile = ""
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal([]byte(src), &data); err != nil {
					log.Fatal(err)
				}
			}

			if err := encoder.Encode(&data); err != nil {
				log.Fatal(err)
			}
		}
	}
}
