package common

import "sync/atomic"

const (
	QueuePolicyBlock      = "block"
	QueuePolicyDropOldest = "drop_oldest"
)

type MessageChan struct {
	in       chan *OctopusEvent
	out      chan *OctopusEvent
	capacity int
	policy   string
	length   atomic.Int64
	dropped  atomic.Int64
}

func NewMessageChan(capacity int) *MessageChan {
	return NewMessageChanWithPolicy(capacity, QueuePolicyBlock)
}

func NewMessageChanWithPolicy(capacity int, policy string) *MessageChan {
	if capacity <= 0 {
		capacity = 1
	}
	if policy != QueuePolicyDropOldest {
		policy = QueuePolicyBlock
	}

	in := make(chan *OctopusEvent, capacity)
	out := make(chan *OctopusEvent, capacity)
	ch := &MessageChan{
		in:       in,
		out:      out,
		capacity: capacity,
		policy:   policy,
	}

	go func() {
		defer close(out)
		buffer := make([]*OctopusEvent, 0, capacity)

		for {
			var outChan chan *OctopusEvent
			var nextVal *OctopusEvent
			if len(buffer) > 0 {
				outChan = out
				nextVal = buffer[0]
			}

			select {
			case val, ok := <-in:
				if !ok {
					for len(buffer) > 0 {
						out <- buffer[0]
						buffer[0] = nil
						buffer = buffer[1:]
						ch.length.Add(-1)
					}
					return
				}

				if len(buffer) < ch.capacity {
					buffer = append(buffer, val)
					ch.length.Add(1)
					continue
				}

				if ch.policy == QueuePolicyDropOldest {
					buffer[0] = nil
					buffer = buffer[1:]
					ch.length.Add(-1)
					ch.dropped.Add(1)

					buffer = append(buffer, val)
					ch.length.Add(1)
					continue
				}

				// Blocking policy: backpressure producer until there is room.
				for len(buffer) >= ch.capacity {
					out <- buffer[0]
					buffer[0] = nil
					buffer = buffer[1:]
					ch.length.Add(-1)
				}
				buffer = append(buffer, val)
				ch.length.Add(1)

			case outChan <- nextVal:
				buffer[0] = nil
				buffer = buffer[1:]
				ch.length.Add(-1)
			}
		}
	}()

	return ch
}

func (ch *MessageChan) In() chan<- *OctopusEvent {
	return ch.in
}

func (ch *MessageChan) Out() <-chan *OctopusEvent {
	return ch.out
}

func (ch *MessageChan) Len() int64 {
	return ch.length.Load()
}

func (ch *MessageChan) Dropped() int64 {
	return ch.dropped.Load()
}
