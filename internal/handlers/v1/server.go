package v1

// ServerImpl implements the V1 API handlers for the photos-ng application.
// It contains the business logic for handling HTTP requests and responses
// for all V1 endpoints including albums, media, and timeline.
type ServerImpl struct{}

// NewServer creates and returns a new instance of ServerImpl.
// This constructor initializes the server implementation that will handle
// all V1 API requests for the photos-ng application.
func NewServer() *ServerImpl {
	return &ServerImpl{}
}
