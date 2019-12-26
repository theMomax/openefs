package metadata

import "time"

// Metadata is an Identifiable object, that can aggregate information by adding
// more Metadata.
type Metadata interface {
	Identifiable
	Append(Metadata) Metadata
}

// Identifiable has a unique identifier. The Identifiable with the latest Time
// should have the highest ID.
type Identifiable interface {
	ID() uint64
	Time() time.Time
}

// Basic describes basic properties of an update and the model it was created of.
type Basic struct {
	// Timestamp refers to the time where the update was received by this
	// system.
	Timestamp time.Time
	// Identifier is a unique identifier. The latest update should have the
	// highest Identifier.
	Identifier uint64
	// Details stores appended information.
	Details Metadata
}

// Append adds detail to Basic metadata. It returns latest of b and data.
func (b *Basic) Append(data Metadata) Metadata {
	if data.ID() > b.Identifier {
		return data.Append(b)
	}

	if b.Details == nil {
		b.Details = data
	} else {
		b.Details.Append(data)
	}
	return b
}

// ID returns the unique identifier.
func (b *Basic) ID() uint64 {
	return b.Identifier
}

// Time returns the time, where the associated data was received by the system.
func (b *Basic) Time() time.Time {
	return b.Timestamp
}
