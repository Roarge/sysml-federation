package capacity

import (
	"context"
	"encoding/json"
)

// PopulatePartRequires copies the representation the router sent into the entity by one JSON round trip, so the nested lists arrive typed.
func (ec *executionContext) PopulatePartRequires(ctx context.Context, entity *Part, reps map[string]any) error { //nointerface:allow
	b, err := json.Marshal(reps)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, entity)
}

// PopulateRequirementRequires does the same for a requirement, subject included.
func (ec *executionContext) PopulateRequirementRequires(ctx context.Context, entity *Requirement, reps map[string]any) error { //nointerface:allow
	b, err := json.Marshal(reps)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, entity)
}
