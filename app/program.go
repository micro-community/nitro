// Package app encapsulates the client, server and other interfaces to provide a complete dapp
package app

// Program is an interface for distributed application programming
type Program interface {
	// Set the current application name
	Name(string)
	// Execute a function in a remote program
	Execute(prog, fn string, req, rsp interface{}) error
	// Broadcast an event to subscribers
	Broadcast(event string, msg interface{}) error
	// Register a function e.g a public Go struct/method with signature func(context.Context, *Request, *Response) error
	Register(fn interface{}) error
	// Subscribe to broadcast events. Signature is public Go func or struct with signature func(context.Context, *Message) error
	Subscribe(event string, fn interface{}) error
	// Run the application program
	Run() error
}
