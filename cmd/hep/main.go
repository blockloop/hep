package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/blockloop/hep/internal/config"
	"github.com/blockloop/hep/internal/log"
	"github.com/blockloop/hep/internal/parse"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Error.Fatalf("configuration failure: %v\n", err)
	}

	log.Verbose(cfg.Verbose)

	r, err := parse.ParseCommand(os.Args[1:]...)
	if err != nil {
		log.Error.Fatalf("failed to parse command: %s", err)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error.Fatalf("failed to execute request: %s", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error.Fatalf("failed to read response body: %v", err)
	}

	for k, v := range res.Header {
		fmt.Fprintf(os.Stderr, "%s: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println()
	fmt.Println(string(body))
}
