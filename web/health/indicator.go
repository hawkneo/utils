package health

import "fmt"

type Status string

const (
	// Unknown indicating that the component or subsystem is in an unknown state.
	Unknown Status = "UNKNOWN"
	// Up indicating that the component or subsystem is functioning as expected.
	Up Status = "UP"
	// Down indicating that the component or subsystem has suffered an unexpected failure.
	Down Status = "DOWN"
)

var (
	_ Indicator = (*CompositeIndicator)(nil)
)

func (status Status) Merge(other Status) Status {
	if status == Down || other == Down {
		return Down
	}
	if status == Unknown || other == Unknown {
		return Unknown
	}
	return Up
}

type Health struct {
	Status  Status         `json:"status"`
	Details map[string]any `json:"details"`
}

func NewUpHealth() Health {
	return Health{
		Status:  Up,
		Details: make(map[string]any),
	}
}

func NewDownHealth(err error) Health {
	return Health{
		Status:  Down,
		Details: map[string]any{"error": fmt.Sprintf("%v", err)},
	}
}

func NewUnknownHealth(err error) Health {
	return Health{
		Status:  Unknown,
		Details: map[string]any{"error": fmt.Sprintf("%v", err)},
	}
}

// Merge merges the health of the given indicator into the current health.
func (health Health) Merge(name string, other Health) Health {
	health.Status = health.Status.Merge(other.Status)
	health.Details[name] = other.Details
	return health
}

// Indicator is a component that can be checked for its health.
type Indicator interface {
	// Name returns the name of the indicator.
	Name() string
	// Health returns the health of the indicator.
	Health() Health
}

// CompositeIndicator is an indicator that is composed of other indicators.
type CompositeIndicator struct {
	Indicators []Indicator
}

func (indicator *CompositeIndicator) Name() string {
	return "CompositeIndicator"
}

func (indicator *CompositeIndicator) Health() Health {
	health := Health{
		Status:  Up,
		Details: make(map[string]any),
	}

	for _, each := range indicator.Indicators {
		func(indicator Indicator) {
			defer func() {
				if r := recover(); r != nil {
					health = health.Merge(indicator.Name(), Health{
						Status:  Unknown,
						Details: map[string]any{"error": fmt.Sprintf("%v", r)},
					})
				}
			}()

			health = health.Merge(indicator.Name(), indicator.Health())
		}(each)
	}
	return health
}
