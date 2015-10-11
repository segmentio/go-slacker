
# Slacker

 Slack slash command http.Handler.

```go
slack := slacker.New()

slack.HandleFunc("hello", "foo", func(cmd *slacker.Command) error {
  fmt.Fprint(cmd, "Hello")
  fmt.Fprint(cmd, " World")
  return nil
})

slack.HandleFunc("boom", "foo", func(cmd *slacker.Command) error {
  return fmt.Errorf("something exploded")
})

slack.HandleFunc("deploy", "foo", func(cmd *slacker.Command) error {
  fmt.Fprintf(cmd, "Deploying %q", cmd.Text)
  return nil
})

log.Fatal(http.ListenAndServe(":8080", slack))
```


## Testing Locally
Use the [slacker-cli](https://github.com/segmentio/slacker-cli) tool, which
spins up a local chat room that can talk to your Slack custom slash command
server.


## Usage

#### type Command

```go
type Command struct {
  Name        string
  Text        string
  Token       string
  UserID      string
  UserName    string
  ChannelID   string
  ChannelName string
}
```

Command details sent by Slack.

#### type Handler

```go
type Handler interface {
  HandleCommand(cmd *Command) (string, error)
}
```

Handler.

#### type HandlerFunc

```go
type HandlerFunc func(cmd *Command) (string, error)
```

HandlerFunc convenience type.

#### func (HandlerFunc) HandleCommand

```go
func (h HandlerFunc) HandleCommand(cmd *Command) (string, error)
```
HandleCommand invokes itself.

#### type Slacker

```go
type Slacker struct {
  sync.Mutex
}
```

Slacker handles HTTP requests and command dispatching.

#### func  New

```go
func New() *Slacker
```
New slacker.

#### func (*Slacker) Handle

```go
func (s *Slacker) Handle(name, token string, handler Handler)
```
Handle registers `handler` for command `name` with `token`.

#### func (*Slacker) HandleFunc

```go
func (s *Slacker) HandleFunc(name, token string, handler func(*Command) (string, error))
```
HandleFunc registers `handler` function for command `name` with `token`.

#### func (*Slacker) ServeHTTP

```go
func (s *Slacker) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP handles slash command requests.

#### func (*Slacker) ValidToken

```go
func (s *Slacker) ValidToken(command, token string) bool
```
ValidToken validates the given `token` for the command. Returns `false` if the command is not registered.
