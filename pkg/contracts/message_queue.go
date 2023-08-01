package contracts

type MessageQueue interface {
	Publish(payload []byte) error
	Setup()
}

type MessagePayload struct {
	UserID   int64
	Title    string
	Body     string
	ImageUrl string
}
