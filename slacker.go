package slacker

import (
	"bytes"
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

// Bytes returns the string written to the command,
// primarily used for testing.
func (c *Command) String() string {
	return c.buf.String()
}

// Slacker handles HTTP requests and command dispatching.
type Slacker struct {
	handlers map[string]Handler
	sync.Mutex
	tokens map[string]bool
	debug  bool
}

// New slacker with valid `tokens`.
func New(tokens []string, debug bool) *Slacker {
	return &Slacker{
		handlers: make(map[string]Handler),
		tokens:   toMap(tokens),
		debug:    debug,
	}
}

// ValidToken validates the given `token` against the set provided.
func (s *Slacker) ValidToken(token string) bool {
	s.Lock()
	defer s.Unlock()
	if s.debug {
		return true
	}
	return s.tokens[token]
}

// Handle registers `handler` for command `name`.
func (s *Slacker) Handle(name string, handler Handler) {
	s.handlers[name] = handler
}

// HandleFunc registers `handler` function for command `name`.
func (s *Slacker) HandleFunc(name string, handler func(*Command) error) {
	s.Handle(name, HandlerFunc(handler))
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

	if !s.ValidToken(cmd.Token) {
		log.Printf("[error] invalid token %q", cmd.Token)
		http.Error(w, "Invalid token", 401)
		return
	}

	log.Printf("[info] received %s %q from %s in %s", cmd.Name, cmd.Text, cmd.UserName, cmd.ChannelName)

	h, ok := s.handlers[cmd.Name]

	if !ok {
		log.Printf("[error] invalid command %q", cmd.Name)
		http.Error(w, "Invalid command", 400)
		return
	}

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
