package service

// Code is a code that uniquely identifies a Service
type Code = string

// Name is the human-readable/friedly name of a Service
type Name = string

// EndpointCode is a code that uniquely identifies an Endpoint. It need not be globally unique, only unique within that one Service
type EndpointCode = string

// EndpointName is the human-readable/friedly name of an Endpoint
type EndpointName = string

// ServiceEndpointDependencies is a map of Service Endpoints you depend on. Keys are the Service Code, for each service
// you depend on, and values are the specific EndpointCodes within said Service that you depend on
type ServiceEndpointDependencies = map[Code][]EndpointCode

// Service represents a service
type Service struct {
	// ID uniquely identifies this service
	ID *int64
	// Code is a unique code for the service. For example, "location_tracking" for a location tracking service
	Code Code
	// Name is a friendly name for the service. For example, "Location Tracking" for a location tracking service
	Name Name
	// Endpoints is a list of Endpoints that the Service has
	Endpoints []Endpoint
}

// Endpoint represents an Endpoint of a Service
type Endpoint struct {
	// ID uniquely identifies this endpoint
	ID *int64
	// Code is a unique code for the Endpoint. Doesn't have to be globally unique, just unique per service. For example,
	// "POST /locations" or "GET /locations/:id"
	Code EndpointCode
	// Name is a friendly name for the Endpoint. For example, "Create location" or "Get location by id"
	Name EndpointName
	// ServiceEndpointDependencies lists the Service Endpoints you depend on
	ServiceEndpointDependencies ServiceEndpointDependencies
}

// MakeService constructs a Service
func MakeService(
	id *int64,
	code Code,
	name Name,
	endpoints []Endpoint,
) Service {
	return Service{
		ID:        id,
		Code:      code,
		Name:      name,
		Endpoints: endpoints,
	}
}
