package main

import "github.com/segmentio/go-slacker"
import "github.com/tj/docopt"
import "net/http"
import "fmt"
import "log"

var Version = "0.0.1"

const Usage = `
  Usage:
    slacker [--bind addr] [--token token]...
    slacker -h | --help
    slacker --version

  Options:
    --bind addr         bind address [default: :3000]
    -t, --token token   valid token
    -h, --help          output help information
    -v, --version       output version

`

func main() {
	args, err := docopt.Parse(Usage, nil, true, Version, false)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	addr := args["--bind"].(string)
	tokens := args["--token"].([]string)

	log.Printf("[info] starting slacker %s", Version)
	slack := slacker.New(tokens)

	slack.HandleFunc("hello", func(cmd *slacker.Command) (string, error) {
		return "World", nil
	})

	slack.HandleFunc("boom", func(cmd *slacker.Command) (string, error) {
		return "", fmt.Errorf("something exploded")
	})

	slack.HandleFunc("deploy", func(cmd *slacker.Command) (string, error) {
		return "Deploying!", nil
	})

	log.Fatal(http.ListenAndServe(addr, slack))
}
