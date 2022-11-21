package contracts

type MessageQueue interface {
	Publish(title string) error
	Setup()
}
