package service

// Code is a code that uniquely identifies a Service
type Code = string

// Name is the human-readable/friedly name of a Service
type Name = string

// EndpointCode is a code that uniquely identifies an Endpoint. It need not be globally unique, only unique within that one Service
type EndpointCode = string

// EndpointName is the human-readable/friedly name of an Endpoint
type EndpointName = string

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
	// Dependencies is a map of Dependencies for a given Endpoint. Keys are service codes, values are lists of endpoint codes
	Dependencies map[Code][]EndpointCode
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
