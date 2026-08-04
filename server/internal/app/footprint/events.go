package footprint

import "time"

type JourneyCreated struct {
	ID int64
	At time.Time
}

func (e JourneyCreated) Name() string          { return "footprint.journey.created" }
func (e JourneyCreated) OccurredAt() time.Time { return e.At }

type JourneyUpdated struct {
	ID int64
	At time.Time
}

func (e JourneyUpdated) Name() string          { return "footprint.journey.updated" }
func (e JourneyUpdated) OccurredAt() time.Time { return e.At }

type JourneyDeleted struct {
	ID int64
	At time.Time
}

func (e JourneyDeleted) Name() string          { return "footprint.journey.deleted" }
func (e JourneyDeleted) OccurredAt() time.Time { return e.At }
