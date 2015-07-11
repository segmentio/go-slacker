package slacker

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// Handler.
type Handler interface {
	HandleCommand(cmd *Command) error
}

// HandlerFunc convenience type.
type HandlerFunc func(cmd *Command) error

// HandleCommand invokes itself.
func (h HandlerFunc) HandleCommand(cmd *Command) error {
	return h(cmd)
}

// Command details sent by Slack.
type Command struct {
	Name        string
	Text        string
	Token       string
	UserID      string
	UserName    string
	ChannelID   string
	ChannelName string
	buf         bytes.Buffer
}

// Write to the internal bytes.Buffer.
func (c *Command) Write(p []byte) (int, error) {
	return c.buf.Write(p)
}

// Bytes returns the bytes written to the command,
// primarily used for testing.
func (c *Command) Bytes() []byte {
	return c.buf.Bytes()
}

// String returns the string written to the command,
// primarily used for testing.
func (c *Command) String() string {
	return c.buf.String()
}

// Slacker handles HTTP requests and command dispatching.
type Slacker struct {
	handlers map[string]Handler // maps a command to its handler.
	tokens   map[string]string  // maps a command to its token.
	sync.Mutex
}

// New slacker.
func New(tokens []string) *Slacker {
	return &Slacker{
		handlers: make(map[string]Handler),
	}
}

// ValidToken validates the given `token` for the given `command`.
func (s *Slacker) ValidToken(command, token string) bool {
	s.Lock()
	defer s.Unlock()

	// Under normal execution, we would have already validated whether the command
	// exists or not. But since this an exported function, we must validate that
	// the command does indeed exist.
	t, ok := s.tokens[command]
	if ok {
		return t == token
	}

	return false
}

// Handle registers `handler` for command `name` with `token`.
func (s *Slacker) Handle(name, token string, handler Handler) {
	s.handlers[name] = handler
	s.tokens[name] = token
}

// HandleFunc registers `handler` function for command `name` with `token`.
func (s *Slacker) HandleFunc(name, token string, handler func(*Command) error) {
	s.Handle(name, token, HandlerFunc(handler))
}

// ServeHTTP handles slash command requests.
func (s *Slacker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("[error] parsing form: %s", err)
		http.Error(w, "Invalid request body", 400)
		return
	}

	command := r.Form.Get("command")

	if command == "" {
		http.Error(w, "command required", 400)
		return
	}

	cmd := &Command{
		Name:        command[1:],
		Text:        r.Form.Get("text"),
		Token:       r.Form.Get("token"),
		UserID:      r.Form.Get("user_id"),
		UserName:    r.Form.Get("user_name"),
		ChannelID:   r.Form.Get("channel_id"),
		ChannelName: r.Form.Get("channel_name"),
	}

	h, ok := s.handlers[cmd.Name]
	if !ok {
		log.Printf("[error] invalid command %q", cmd.Name)
		http.Error(w, "Invalid command", 400)
		return
	}

	if !s.ValidToken(cmd.Name, cmd.Token) {
		log.Printf("[error] invalid token %q for command %q", cmd.Token, cmd.Name)
		http.Error(w, fmt.Sprintf("Invalid token %q for command %q", cmd.Token, cmd.Name), 401)
		return
	}

	log.Printf("[info] received %s %q from %s in %s", cmd.Name, cmd.Text, cmd.UserName, cmd.ChannelName)

	err = h.HandleCommand(cmd)
	if err != nil {
		log.Printf("[error] handling command: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = io.Copy(w, &cmd.buf)
	if err != nil {
		log.Printf("[error] writing: %s", err)
	}
}

// Map from string slice.
func toMap(s []string) map[string]bool {
	m := make(map[string]bool)
	for _, k := range s {
		m[k] = true
	}
	return m
}
