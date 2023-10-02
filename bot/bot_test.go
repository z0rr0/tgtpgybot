package bot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/z0rr0/tgtpgybot/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{Offline: true}

	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if b.bot == nil {
		t.Fatal("bot is nil")
	}

	if b.cfg != cfg {
		t.Fatal("config is not equal")
	}
}

func TestBot_Start(t *testing.T) {
	cfg := &config.Config{Offline: true}

	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	sigChan := make(chan os.Signal)

	go func() {
		time.Sleep(time.Second)
		sigChan <- os.Interrupt
	}()

	b.Start(sigChan)
	b.Stop()
}

type testContext struct{}

func (m *testContext) Bot() *telebot.Bot                                 { return nil }
func (m *testContext) Update() telebot.Update                            { return telebot.Update{} }
func (m *testContext) Message() *telebot.Message                         { return &telebot.Message{ID: 2, Text: "test"} }
func (m *testContext) Callback() *telebot.Callback                       { return nil }
func (m *testContext) Query() *telebot.Query                             { return nil }
func (m *testContext) InlineResult() *telebot.InlineResult               { return nil }
func (m *testContext) ShippingQuery() *telebot.ShippingQuery             { return nil }
func (m *testContext) PreCheckoutQuery() *telebot.PreCheckoutQuery       { return nil }
func (m *testContext) ChatMember() *telebot.ChatMemberUpdate             { return nil }
func (m *testContext) ChatJoinRequest() *telebot.ChatJoinRequest         { return nil }
func (m *testContext) Poll() *telebot.Poll                               { return nil }
func (m *testContext) PollAnswer() *telebot.PollAnswer                   { return nil }
func (m *testContext) Migration() (int64, int64)                         { return 0, 0 }
func (m *testContext) Sender() *telebot.User                             { return &telebot.User{ID: 1, Username: "test"} }
func (m *testContext) Chat() *telebot.Chat                               { return nil }
func (m *testContext) Recipient() telebot.Recipient                      { return nil }
func (m *testContext) Text() string                                      { return "" }
func (m *testContext) Entities() telebot.Entities                        { return nil }
func (m *testContext) Data() string                                      { return "" }
func (m *testContext) Args() []string                                    { return []string{"arg1", "arg2"} }
func (m *testContext) Send(interface{}, ...interface{}) error            { return nil }
func (m *testContext) SendAlbum(telebot.Album, ...interface{}) error     { return nil }
func (m *testContext) Reply(interface{}, ...interface{}) error           { return nil }
func (m *testContext) Forward(telebot.Editable, ...interface{}) error    { return nil }
func (m *testContext) ForwardTo(telebot.Recipient, ...interface{}) error { return nil }
func (m *testContext) Edit(interface{}, ...interface{}) error            { return nil }
func (m *testContext) EditCaption(string, ...interface{}) error          { return nil }
func (m *testContext) EditOrSend(interface{}, ...interface{}) error      { return nil }
func (m *testContext) EditOrReply(interface{}, ...interface{}) error     { return nil }
func (m *testContext) Delete() error                                     { return nil }
func (m *testContext) DeleteAfter(time.Duration) *time.Timer             { return nil }
func (m *testContext) Notify(telebot.ChatAction) error                   { return nil }
func (m *testContext) Ship(...interface{}) error                         { return nil }
func (m *testContext) Accept(...string) error                            { return nil }
func (m *testContext) Respond(...*telebot.CallbackResponse) error        { return nil }
func (m *testContext) Answer(*telebot.QueryResponse) error               { return nil }
func (m *testContext) Set(string, interface{})                           {}
func (m *testContext) Get(string) interface{}                            { return nil }

func TestBotRootHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"result":{"message":{"role":"Ассистент","text":"Меня зовут Алиса"},"num_tokens":"20"}}`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		Offline: true,
		Timeout: config.TimeDuration{Duration: 5 * time.Second},
		Chat:    config.Chat{APIKey: "test-key", URL: s.URL, Client: s.Client()},
	}

	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	c := &testContext{}
	if err = b.rootHandler(c); err != nil {
		t.Fatal(err)
	}
}

func TestDurationMiddleware(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"id":"test","created":1677652288,` +
			`"data":[{"url":"https://127.0.0.1/test1"},{"url":"https://127.0.0.1/test2"}]}`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		Offline: true,
		Timeout: config.TimeDuration{Duration: 5 * time.Second},
		Chat:    config.Chat{APIKey: "test-key", URL: s.URL, Client: s.Client()},
	}

	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	c := &testContext{}
	firstHandler := durationMiddleware()
	secondHandler := firstHandler(b.rootHandler)

	if err = secondHandler(c); err != nil {
		t.Fatal(err)
	}
}

func TestFailedDurationMiddleware(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test", http.StatusInternalServerError)
	}))
	defer s.Close()

	cfg := &config.Config{
		Offline: true,
		Timeout: config.TimeDuration{Duration: 5 * time.Second},
		Chat:    config.Chat{APIKey: "test-key", URL: s.URL, Client: s.Client()},
	}

	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	c := &testContext{}
	firstHandler := durationMiddleware()
	secondHandler := firstHandler(b.rootHandler)

	if err = secondHandler(c); err != nil {
		t.Fatal(err)
	}
}
