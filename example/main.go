package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/segmentio/go-slacker"
	"github.com/tj/docopt"
)

const (
	version = "0.0.1"

	usage = `
  Usage:
    slacker [--bind addr] [--token token]
    slacker -h | --help
    slacker --version

  Options:
    --bind addr         bind address [default: :3000]
    -t, --token token   valid token
    -h, --help          output help information
    -v, --version       output version`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	addr := args["--bind"].(string)
	token := args["--token"].(string)

	log.Printf("[info] starting slacker %s", version)
	slack := slacker.New()

	slack.HandleFunc("hello", token, func(w io.Writer, cmd *slacker.Command) error {
		fmt.Fprint(w, "Hello")
		fmt.Fprint(w, " World")
		return nil
	})

	slack.HandleFunc("boom", token, func(w io.Writer, cmd *slacker.Command) error {
		return fmt.Errorf("something exploded")
	})

	slack.HandleFunc("deploy", token, func(w io.Writer, cmd *slacker.Command) error {
		fmt.Fprintf(w, "Deploying %q", cmd.Text)
		return nil
	})

	log.Fatal(http.ListenAndServe(addr, slack))
}
