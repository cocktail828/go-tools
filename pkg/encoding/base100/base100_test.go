package base100

import "testing"

func TestInvalidInput(t *testing.T) {
	if _, err := NewEncoding().Decode([]byte("aaaa")); err != ErrInvalidData {
		t.Errorf("Expected ErrInvalidData but got %v", err)
	}

	if _, err := NewEncoding().Decode([]byte("aaa")); err != ErrInvalidLength {
		t.Errorf("Expected ErrInvalidLength but got %v", err)
	}
}

var tests = []struct {
	name  string
	text  string
	emoji string
}{
	{
		"ASCII",
		"hello",
		"👟👜👣👣👦",
	},
	{
		"Cyrillic",
		"РАШ Бэ",
		"📇💗📇💇📇💟🐗📇💈📈💄",
	},
	{
		"HelloUnicode",
		"Hello, 世界",
		"🐿👜👣👣👦🐣🐗📛💯💍📞💌💃",
	},
}

func TestDecode(t *testing.T) {
	for _, test := range tests {
		res, err := NewEncoding().Decode([]byte(test.emoji))
		if err != nil {
			t.Errorf("%v: Unexpected error: %v", test.name, err)
		}

		if string(res) != test.text {
			t.Errorf("%v: Expected to get '%v', got '%v'", test.name, test.text, string(res))
		}
	}
}

func TestEncode(t *testing.T) {
	for _, test := range tests {
		res, err := NewEncoding().Encode([]byte(test.text))
		if err != nil {
			t.Errorf("%v: Unexpected error: %v", test.name, err)
		}

		if string(res) != test.emoji {
			t.Errorf("%v: Expected to get '%v', got '%v'", test.name, test.emoji, res)
		}
	}
}

func TestFlow(t *testing.T) {
	text := []byte("the quick brown fox 😂😂👌👌👌 over the lazy dog привет")

	encoded, err := NewEncoding().Encode(text)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	res, err := NewEncoding().Decode(encoded)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(res) != string(text) {
		t.Errorf("Expected to get '%v', got '%v'", string(text), string(res))
	}
}
