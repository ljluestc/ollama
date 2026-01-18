package template

import (
	"bytes"
	"os"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestGraniteFIM(t *testing.T) {
	bts, err := os.ReadFile("granite-instruct.gotmpl")
	if err != nil {
		t.Fatal(err)
	}

	tmpl, err := Parse(string(bts))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("fim", func(t *testing.T) {
		values := Values{
			Prompt: "def add(",
			Suffix: "    return x + y",
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, values); err != nil {
			t.Fatal(err)
		}


		expected := "<|fim_prefix|>def add(<|fim_suffix|>    return x + y<|fim_middle|>"
		if bytes.TrimSpace(b.Bytes()) != nil && string(bytes.TrimSpace(b.Bytes())) != expected {
             // Doing exact string comparison but handling expected newline if it's just one
        }
        
        // Let's just use string comparison but trim the got string if it has suffix \n
        got := b.String()
        if len(got) > 0 && got[len(got)-1] == '\n' {
            got = got[:len(got)-1]
        }

		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("chat", func(t *testing.T) {
		values := Values{
			Messages: []api.Message{
				{Role: "user", Content: "Hello"},
			},
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, values); err != nil {
			t.Fatal(err)
		}

		expected := "Question:\nHello\n\nAnswer:\n"
		if b.String() != expected {
			t.Errorf("expected %q, got %q", expected, b.String())
		}
	})
}
