package commonlog

import (
	"fmt"

	"github.com/tliron/kutil/util"
)

//
// Message
//

type Message interface {
	// Sets a value on the message and returns the same message
	// object.
	//
	// These keys are often specially supported:
	//
	// "message": the base text of the message
	// "scope": the scope of the message
	Set(key string, value any) Message

	// Sends the message to the backend
	Send()
}

// Calls [Message.Set] on a provided sequence of key-value pairs.
// Thus keysAndValues must have an even length.
//
// Non-string keys are converted to strings using [util.ToString].
func SetMessageKeysAndValue(message Message, keysAndValues ...any) {
	length := len(keysAndValues)

	if length == 0 {
		return
	}

	if length%2 != 0 {
		panic(fmt.Sprintf("CommonLog message keysAndValues does not have an even number of arguments: %d", length))
	}

	for index := 0; index < length; index += 2 {
		key := util.ToString(keysAndValues[index])
		value := keysAndValues[index+1]
		message.Set(key, value)
	}
}

//
// UnstructuredMessage
//

type SendUnstructuredMessageFunc func(message string)

// Convenience type for implementing unstructured backends. Converts a structured
// message to an unstructured string.
type UnstructuredMessage struct {
	prefix  string
	message string
	suffix  string
	send    SendUnstructuredMessageFunc
}

func NewUnstructuredMessage(send SendUnstructuredMessageFunc) *UnstructuredMessage {
	return &UnstructuredMessage{
		send: send,
	}
}

// ([Message] interface)
func (self *UnstructuredMessage) Set(key string, value any) Message {
	switch key {
	case "message":
		self.message = util.ToString(value)

	case "scope":
		self.prefix = "{" + util.ToString(value) + "}"

	default:
		if len(self.suffix) > 0 {
			self.suffix += ", "
		}
		self.suffix += key + "=" + util.ToString(value)
	}

	return self
}

// ([Message] interface)
func (self *UnstructuredMessage) Send() {
	message := self.prefix

	if len(self.message) > 0 {
		if len(message) > 0 {
			message += " "
		}
		message += self.message
	}

	if len(self.suffix) > 0 {
		if len(message) > 0 {
			message += " "
		}
		message += self.suffix
	}

	self.send(message)
}
