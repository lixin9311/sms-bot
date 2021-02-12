package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// SMS holds the schema definition for the SMS entity.
type SMS struct {
	ent.Schema
}

// Fields of the SMS.
func (SMS) Fields() []ent.Field {
	return []ent.Field{
		field.String("number").Immutable(),
		field.Bytes("data").Immutable(),
		field.String("text").Immutable(),
		field.Time("discharge_timestamp").Immutable(),
		// for sender
		field.Enum("delivery_state").Values("RECEIVED", "UNKNOWN", "UNCONFIRMED").Default("UNCONFIRMED"),
	}
}

// Edges of the SMS.
func (SMS) Edges() []ent.Edge {
	return nil
}

// Mixin of the SMS.
func (SMS) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
