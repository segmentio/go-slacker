package slacker_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"testing"

	"github.com/bmizerany/assert"
	"github.com/segmentio/go-slacker"
)

func TestParsesForm(t *testing.T) {
	slack := slacker.New()
	helloC := make(chan *slacker.Command)
	slack.HandleFunc("hello", "foo", func(cmd *slacker.Command) error {
		go func() {
			helloC <- cmd
		}()
		return nil
	})
	ts := httptest.NewServer(slack)
	defer ts.Close()

	values := url.Values{}
	values.Add("command", "/hello")
	values.Add("text", "hello world")
	values.Add("token", "foo")
	values.Add("user_id", "a")
	values.Add("user_name", "b")
	values.Add("channel_id", "c")
	values.Add("channel_name", "d")

	res, err := http.PostForm(ts.URL, values)
	if err != nil {
		t.Fatalf("could not post request with error: %s", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("server responded with %s", res.Status)
	}

	cmd := <-helloC
	assert.Equal(t, "hello", cmd.Name)
	assert.Equal(t, "hello world", cmd.Text)
	assert.Equal(t, "foo", cmd.Token)
	assert.Equal(t, "a", cmd.UserID)
	assert.Equal(t, "b", cmd.UserName)
	assert.Equal(t, "c", cmd.ChannelID)
	assert.Equal(t, "d", cmd.ChannelName)
}

// Make a post request to the given url with the given values and verify the response code and body.
func testResponse(t *testing.T, url string, values url.Values, expectedStatus int, expectedBody string) {
	resp, err := http.PostForm(url, values)
	if err != nil {
		t.Fatalf("could not post request with error: %s", err)
	}
	assert.Equal(t, expectedStatus, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could ready body with errror: %s", err)
	}
	assert.Equal(t, expectedBody+"\n", string(body))
}

func TestCommandIsRequired(t *testing.T) {
	slack := slacker.New()
	ts := httptest.NewServer(slack)
	defer ts.Close()

	values := url.Values{}
	values.Add("command", "")

	testResponse(t, ts.URL, values, 400, "command required")
}

func TestFailsForInvalidCommand(t *testing.T) {
	slack := slacker.New()
	ts := httptest.NewServer(slack)
	defer ts.Close()

	values := url.Values{}
	values.Add("command", "/hello")

	testResponse(t, ts.URL, values, 400, "Invalid command")
}

func TestFailsForInvalidToken(t *testing.T) {
	slack := slacker.New()
	slack.HandleFunc("hello", "foo", func(cmd *slacker.Command) error {
		return nil
	})
	ts := httptest.NewServer(slack)
	defer ts.Close()

	values := url.Values{}
	values.Add("command", "/hello")
	values.Add("token", "non-foo")

	testResponse(t, ts.URL, values, 401, "Invalid token \"non-foo\" for command \"hello\"")
}

func TestFailsWhenHandlerErrors(t *testing.T) {
	slack := slacker.New()
	slack.HandleFunc("hello", "foo", func(cmd *slacker.Command) error {
		return fmt.Errorf("test error")
	})
	ts := httptest.NewServer(slack)
	defer ts.Close()

	values := url.Values{}
	values.Add("command", "/hello")
	values.Add("token", "foo")

	testResponse(t, ts.URL, values, 500, "test error")
}

func TestValidatesToken(t *testing.T) {
	slack := slacker.New()
	empty := func(cmd *slacker.Command) error {
		return nil
	}
	slack.HandleFunc("foo", "bar", empty)
	slack.HandleFunc("qaz", "qux", empty)

	// Correct tokens are validated.
	assert.Equal(t, true, slack.ValidToken("foo", "bar"))
	assert.Equal(t, true, slack.ValidToken("qaz", "qux"))
	// Tokens are only valid for specific commands.
	assert.Equal(t, false, slack.ValidToken("foo", "qux"))
	assert.Equal(t, false, slack.ValidToken("qaz", "bar"))
	// Invalid commands are not validated.
	assert.Equal(t, false, slack.ValidToken("foobar", ""))
}
