package api

type Event interface {
	IsEvent(other interface{}) bool
	Marshal() ([]byte, error)
}

type EventService interface {
	EventFeed() <-chan Event
}
