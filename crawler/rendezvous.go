package crawler

type Receiver[V any] chan<- V
type Sender[V any] <-chan V

// receiver and sender should be buffered channel with length of 1
func Rendezvous[V any](receiver Receiver[V], sender Sender[V]) func(sendData V) Sender[V] {
	return func(sendData V) Sender[V] {
		receiver <- sendData

		return sender
	}
}
