package message

import (
	"github.com/cocktail828/go-tools/message/messagepb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	*messagepb.Message
	private interface{}
}

func New(body []byte) (*Message, error) {
	rawmsg := messagepb.Message{}
	if err := proto.Unmarshal(body, &rawmsg); err != nil {
		return nil, err
	}
	return &Message{Message: &rawmsg}, nil
}

func (m *Message) Private() interface{} {
	return m.private
}

func (m *Message) Parse(f func([]byte) (interface{}, error)) error {
	if f == nil {
		return errors.Errorf("invalid unmarshaller for sub:%v", m.Sub)
	}

	if m.private != nil {
		return nil
	}

	v, err := f(m.GetData())
	if err == nil {
		m.private = v
	}
	return err
}
