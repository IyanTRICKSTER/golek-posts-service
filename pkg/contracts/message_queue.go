package contracts

type MessageQueue interface {
	Publish(payload []byte) error
	Setup()
}

type MessagePayload struct {
	Title    string
	Body     string
	ImageUrl string
}
