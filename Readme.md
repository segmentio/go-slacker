
# Slacker

 Slack slash command http.Handler.

```go
slack := slacker.New(tokens)

slack.HandleFunc("hello", func(cmd *slacker.Command) error {
  fmt.Fprint(cmd, "Hello")
  fmt.Fprint(cmd, " World")
  return nil
})

slack.HandleFunc("boom", func(cmd *slacker.Command) error {
  return fmt.Errorf("something exploded")
})

slack.HandleFunc("deploy", func(cmd *slacker.Command) error {
  fmt.Fprintf(cmd, "Deploying %q", cmd.Text)
  return nil
})
```

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
func New(tokens []string) *Slacker
```
New slacker with valid `tokens`.

#### func (*Slacker) Handle

```go
func (s *Slacker) Handle(name string, handler Handler)
```
Handle registers `handler` for command `name`.

#### func (*Slacker) HandleFunc

```go
func (s *Slacker) HandleFunc(name string, handler func(*Command) (string, error))
```
HandleFunc registers `handler` function for command `name`.

#### func (*Slacker) ServeHTTP

```go
func (s *Slacker) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP handles slash command requests.

#### func (*Slacker) ValidToken

```go
func (s *Slacker) ValidToken(token string) bool
```
ValidToken validates the given `token` against the set provided.
