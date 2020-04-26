package forms

type (
	Message struct {
		Text  string
		Nonce string
	}
	Topic struct {
		Text  string
		Nonce string
	}
)

func NewMessage(t string) *Message {
	return &Message{
		Text: t,
	}
}

func NewTopic(t string) *Topic {
	return &Topic{
		Text: t,
	}
}
