package mediarecord

import "time"

type MediaRecordCreated struct {
	ID        int64
	Published bool
	At        time.Time
}

func (e MediaRecordCreated) Name() string          { return "media_record.created" }
func (e MediaRecordCreated) OccurredAt() time.Time { return e.At }

type MediaRecordUpdated struct {
	ID        int64
	Published bool
	At        time.Time
}

func (e MediaRecordUpdated) Name() string          { return "media_record.updated" }
func (e MediaRecordUpdated) OccurredAt() time.Time { return e.At }

type MediaRecordDeleted struct {
	ID int64
	At time.Time
}

func (e MediaRecordDeleted) Name() string          { return "media_record.deleted" }
func (e MediaRecordDeleted) OccurredAt() time.Time { return e.At }
