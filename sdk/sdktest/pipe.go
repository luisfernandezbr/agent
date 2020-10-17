package sdktest

import "github.com/pinpt/agent/v4/sdk"

// MockPipe stores sent things inside it
type MockPipe struct {
	// Written is all the models written to this pipe in order
	Written []sdk.Model
	// Closed is set to true every time Close() is called
	Closed bool
	// Flushed is set to true every time Flush() is called
	Flushed bool

	// WriteErr is returned by Write
	WriteErr error
	// FlushErr is returned by Flush
	FlushErr error
	// CloseErr is returned by Close
	CloseErr error
}

var _ sdk.Pipe = (*MockPipe)(nil)

// Write a model back to the output system
func (p *MockPipe) Write(object sdk.Model) error {
	p.Written = append(p.Written, object)
	return p.WriteErr
}

// Flush will tell the pipe to flush any pending data
func (p *MockPipe) Flush() error {
	p.Flushed = true
	return p.FlushErr
}

// Close is called when the integration has completed and no more data will be sent
func (p *MockPipe) Close() error {
	p.Closed = true
	return p.CloseErr
}
