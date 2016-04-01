package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	for _, srcFile := range srcFiles {
		var data interface{}
		var decoder *json.Decoder
		if srcFile == "-" {
			decoder = json.NewDecoder(os.Stdin)
		}
		proceed := true
		for proceed && (srcFile != "-" || decoder.More()) {
			if srcFile == "-" {
				if err := decoder.Decode(&data); err != nil {
					log.Fatal(err)
				}
			} else {
				proceed = false
				src, err := ioutil.ReadFile(srcFile)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal([]byte(src), &data); err != nil {
					log.Fatal(err)
				}
			}

			if result, err := canonicaljson.Marshal(&data); err != nil {
				log.Fatal(err)
			} else {
				fmt.Printf("%s", string(result))
			}
		}
	}
}
