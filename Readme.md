# Slacker

[![GoDoc](https://godoc.org/github.com/segmentio/go-slacker?status.svg)](https://godoc.org/github.com/segmentio/go-slacker)

 Slack slash command `http.Handler`.

```go
slack := slacker.New()

slack.HandleFunc("hello", "<token>", func(w io.Writer, cmd *slacker.Command) error {
  fmt.Fprint(w, "Hello")
  fmt.Fprint(w, " World")
  return nil
})

slack.HandleFunc("boom", "<token>", func(w io.Writer, cmd *slacker.Command) error {
  return fmt.Errorf("something exploded")
})

log.Fatal(http.ListenAndServe(":8080", slack))
```

## Testing Locally
Use the [slacker-cli](https://github.com/segmentio/slacker-cli) tool, which spins up a local chat room that can talk to your Slack custom slash command server.
