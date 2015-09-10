package slacker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// Message details sent to Slack.
type message struct {
	text    string
	channel string
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
	public      bool
	buf         bytes.Buffer
}

// Write to the internal bytes.Buffer.
func (c *Command) Write(p []byte) (int, error) {
	return c.buf.Write(p)
}

// Bytes returns the bytes written to the command.
func (c *Command) Bytes() []byte {
	return c.buf.Bytes()
}

// String returns the string written to the command.
func (c *Command) String() string {
	return c.buf.String()
}

// Public marks the response to the command to be redirected as a public response.
func (c *Command) Public() {
	c.public = true
}

// Slacker handles HTTP requests and command dispatching.
type Slacker struct {
	handlers map[string]Handler // maps a command to its handler.
	tokens   map[string]string  // maps a command to its token.
	webhook  string
	sync.Mutex
}

// New slacker.
func New(webhook string) *Slacker {
	return &Slacker{
		handlers: make(map[string]Handler),
		tokens:   make(map[string]string),
		webhook:  webhook,
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

	if !cmd.public {
		_, err = io.Copy(w, &cmd.buf)
		if err != nil {
			log.Printf("[error] writing: %s", err)
		}
		return
	}

	if s.webhook == "" {
		log.Printf("[error] no webhook specified to post command %s", cmd.Text)
		http.Error(w, "no webhook url specified to post command publicly", http.StatusInternalServerError)
		return
	}

	msg := &message{
		text:    cmd.String(),
		channel: "#" + cmd.ChannelName,
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "error constructing public message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(s.webhook, "application/json", bytes.NewReader(buf))
	if err != nil {
		http.Error(w, "error sending public message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		output, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, "slack rejected public message with "+string(output), http.StatusInternalServerError)
	}
}
