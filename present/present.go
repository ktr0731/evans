// Package present defines a presenter for formatting response types represented by the codec.
package present

// Presenter formats response types returned from the gRPC server for displaying it.
type Presenter interface {
	// Format receives a response v and returns the formatted output as string.
	// Note that v is a type that belongs to the selected IDL.
	//
	// For example, v is a proto.Message if IDL is Protocol Buffers.
	Format(v interface{}) (string, error)
}
