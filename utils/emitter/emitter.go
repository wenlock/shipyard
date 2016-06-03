package emitter

type Emitter struct {
	Msg chan Message
	listeners int
}

type Message struct {
	Category string
	Data     interface{}
}

func NewEmitter() *Emitter {
	emitter := new(Emitter)
	emitter.Msg = make(chan Message)
	return emitter
}

func (emitter *Emitter) WaitForMessage() Message {
	emitter.listeners = emitter.listeners + 1
	return <-emitter.Msg
}

func (emitter *Emitter) BroadcastMessage(category string, data interface{}) {
	msg := Message{
		Category: category,
		Data: data,
	}

	for emitter.listeners != 0 {
		emitter.Msg <- msg
		emitter.listeners = emitter.listeners - 1
	}
}