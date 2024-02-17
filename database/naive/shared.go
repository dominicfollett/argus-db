package naive

import "sync"

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	Lock 	sync.Mutex
	Tokens  int
	Time    string
	// and ?
}

// The message that is passed over the channel
type Message struct {
	Key []byte
	Data *Data
}