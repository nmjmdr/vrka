package howler

type Callback struct {
	Uri     string
	Payload string
}

type Howl interface {
	Add(Callback, uint64) (string, error)
	Del(string) error
	Howls() <-chan Callback
	Start()
	Stop()
}
