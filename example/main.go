package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/segmentio/go-slacker"
	"github.com/tj/docopt"
)

var Version = "0.0.1"

const Usage = `
  Usage:
    slacker [--bind addr] [--token token] [--webhook webhook]
    slacker -h | --help
    slacker --version

  Options:
    --bind addr             bind address [default: :3000]
    -t, --token token       valid token
    -w, --webhook webhook   webhook url
    -h, --help              output help information
    -v, --version           output version
`

func main() {
	args, err := docopt.Parse(Usage, nil, true, Version, false)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	addr := args["--bind"].(string)
	token := args["--token"].(string)
	webhook := args["--webhook"].(string)

	log.Printf("[info] starting slacker %s", Version)
	slack := slacker.New(webhook)

	slack.HandleFunc("hello", token, func(cmd *slacker.Command) error {
		fmt.Fprint(cmd, "Hello")
		fmt.Fprint(cmd, " World")
		return nil
	})

	slack.HandleFunc("boom", token, func(cmd *slacker.Command) error {
		return fmt.Errorf("something exploded")
	})

	slack.HandleFunc("deploy", token, func(cmd *slacker.Command) error {
		cmd.Public()
		fmt.Fprintf(cmd, "Deploying %q", cmd.Text)
		return nil
	})

	log.Fatal(http.ListenAndServe(addr, slack))
}
