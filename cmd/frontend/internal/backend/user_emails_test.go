package backend

import (
	"context"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

func TestSendUserEmailVerificationEmail(t *testing.T) {
	var sent *txemail.Message
	txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
		sent = &message
		return nil
	}
	defer func() { txemail.MockSend = nil }()

	if err := SendUserEmailVerificationEmail(context.Background(), "a@example.com", "c"); err != nil {
		t.Fatal(err)
	}
	if sent == nil {
		t.Fatal("want sent != nil")
	}
	if want := (txemail.Message{
		FromName: "",
		To:       []string{"a@example.com"},
		Template: verifyEmailTemplates,
		Data:     struct{ URL string }{URL: "http://example.com/-/verify-email?code=c"},
	}); !reflect.DeepEqual(*sent, want) {
		t.Errorf("got %+v, want %+v", *sent, want)
	}
}
