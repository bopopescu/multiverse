package controller

// Integration consts to distinct origin of a request.
const (
	IntegrationApplication Integration = iota
	IntegrationBackend
)

// Integration determines the type of integration used for an operation.
type Integration uint8

// Origin information of an operation.
type Origin struct {
	Integration Integration
	UserID      uint64
}

// IsBackend indicates if the origin integration is a backend.
func (o Origin) IsBackend() bool {
	return o.Integration == IntegrationBackend
}