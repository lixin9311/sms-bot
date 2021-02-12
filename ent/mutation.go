// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lixin9311/sms-bot/ent/predicate"
	"github.com/lixin9311/sms-bot/ent/sms"

	"entgo.io/ent"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeSMS = "SMS"
)

// SMSMutation represents an operation that mutates the SMS nodes in the graph.
type SMSMutation struct {
	config
	op                  Op
	typ                 string
	id                  *int
	created_at          *time.Time
	updated_at          *time.Time
	number              *string
	data                *[]byte
	text                *string
	discharge_timestamp *time.Time
	delivery_state      *sms.DeliveryState
	clearedFields       map[string]struct{}
	done                bool
	oldValue            func(context.Context) (*SMS, error)
	predicates          []predicate.SMS
}

var _ ent.Mutation = (*SMSMutation)(nil)

// smsOption allows management of the mutation configuration using functional options.
type smsOption func(*SMSMutation)

// newSMSMutation creates new mutation for the SMS entity.
func newSMSMutation(c config, op Op, opts ...smsOption) *SMSMutation {
	m := &SMSMutation{
		config:        c,
		op:            op,
		typ:           TypeSMS,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withSMSID sets the ID field of the mutation.
func withSMSID(id int) smsOption {
	return func(m *SMSMutation) {
		var (
			err   error
			once  sync.Once
			value *SMS
		)
		m.oldValue = func(ctx context.Context) (*SMS, error) {
			once.Do(func() {
				if m.done {
					err = fmt.Errorf("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().SMS.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withSMS sets the old SMS of the mutation.
func withSMS(node *SMS) smsOption {
	return func(m *SMSMutation) {
		m.oldValue = func(context.Context) (*SMS, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m SMSMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m SMSMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, fmt.Errorf("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID
// is only available if it was provided to the builder.
func (m *SMSMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// SetCreatedAt sets the "created_at" field.
func (m *SMSMutation) SetCreatedAt(t time.Time) {
	m.created_at = &t
}

// CreatedAt returns the value of the "created_at" field in the mutation.
func (m *SMSMutation) CreatedAt() (r time.Time, exists bool) {
	v := m.created_at
	if v == nil {
		return
	}
	return *v, true
}

// OldCreatedAt returns the old "created_at" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldCreatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldCreatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldCreatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreatedAt: %w", err)
	}
	return oldValue.CreatedAt, nil
}

// ResetCreatedAt resets all changes to the "created_at" field.
func (m *SMSMutation) ResetCreatedAt() {
	m.created_at = nil
}

// SetUpdatedAt sets the "updated_at" field.
func (m *SMSMutation) SetUpdatedAt(t time.Time) {
	m.updated_at = &t
}

// UpdatedAt returns the value of the "updated_at" field in the mutation.
func (m *SMSMutation) UpdatedAt() (r time.Time, exists bool) {
	v := m.updated_at
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdatedAt returns the old "updated_at" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldUpdatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldUpdatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldUpdatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdatedAt: %w", err)
	}
	return oldValue.UpdatedAt, nil
}

// ResetUpdatedAt resets all changes to the "updated_at" field.
func (m *SMSMutation) ResetUpdatedAt() {
	m.updated_at = nil
}

// SetNumber sets the "number" field.
func (m *SMSMutation) SetNumber(s string) {
	m.number = &s
}

// Number returns the value of the "number" field in the mutation.
func (m *SMSMutation) Number() (r string, exists bool) {
	v := m.number
	if v == nil {
		return
	}
	return *v, true
}

// OldNumber returns the old "number" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldNumber(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldNumber is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldNumber requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldNumber: %w", err)
	}
	return oldValue.Number, nil
}

// ResetNumber resets all changes to the "number" field.
func (m *SMSMutation) ResetNumber() {
	m.number = nil
}

// SetData sets the "data" field.
func (m *SMSMutation) SetData(b []byte) {
	m.data = &b
}

// Data returns the value of the "data" field in the mutation.
func (m *SMSMutation) Data() (r []byte, exists bool) {
	v := m.data
	if v == nil {
		return
	}
	return *v, true
}

// OldData returns the old "data" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldData(ctx context.Context) (v []byte, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldData is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldData requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldData: %w", err)
	}
	return oldValue.Data, nil
}

// ResetData resets all changes to the "data" field.
func (m *SMSMutation) ResetData() {
	m.data = nil
}

// SetText sets the "text" field.
func (m *SMSMutation) SetText(s string) {
	m.text = &s
}

// Text returns the value of the "text" field in the mutation.
func (m *SMSMutation) Text() (r string, exists bool) {
	v := m.text
	if v == nil {
		return
	}
	return *v, true
}

// OldText returns the old "text" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldText(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldText is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldText requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldText: %w", err)
	}
	return oldValue.Text, nil
}

// ResetText resets all changes to the "text" field.
func (m *SMSMutation) ResetText() {
	m.text = nil
}

// SetDischargeTimestamp sets the "discharge_timestamp" field.
func (m *SMSMutation) SetDischargeTimestamp(t time.Time) {
	m.discharge_timestamp = &t
}

// DischargeTimestamp returns the value of the "discharge_timestamp" field in the mutation.
func (m *SMSMutation) DischargeTimestamp() (r time.Time, exists bool) {
	v := m.discharge_timestamp
	if v == nil {
		return
	}
	return *v, true
}

// OldDischargeTimestamp returns the old "discharge_timestamp" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldDischargeTimestamp(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldDischargeTimestamp is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldDischargeTimestamp requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDischargeTimestamp: %w", err)
	}
	return oldValue.DischargeTimestamp, nil
}

// ResetDischargeTimestamp resets all changes to the "discharge_timestamp" field.
func (m *SMSMutation) ResetDischargeTimestamp() {
	m.discharge_timestamp = nil
}

// SetDeliveryState sets the "delivery_state" field.
func (m *SMSMutation) SetDeliveryState(ss sms.DeliveryState) {
	m.delivery_state = &ss
}

// DeliveryState returns the value of the "delivery_state" field in the mutation.
func (m *SMSMutation) DeliveryState() (r sms.DeliveryState, exists bool) {
	v := m.delivery_state
	if v == nil {
		return
	}
	return *v, true
}

// OldDeliveryState returns the old "delivery_state" field's value of the SMS entity.
// If the SMS object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *SMSMutation) OldDeliveryState(ctx context.Context) (v sms.DeliveryState, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldDeliveryState is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldDeliveryState requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDeliveryState: %w", err)
	}
	return oldValue.DeliveryState, nil
}

// ResetDeliveryState resets all changes to the "delivery_state" field.
func (m *SMSMutation) ResetDeliveryState() {
	m.delivery_state = nil
}

// Op returns the operation name.
func (m *SMSMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (SMS).
func (m *SMSMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *SMSMutation) Fields() []string {
	fields := make([]string, 0, 7)
	if m.created_at != nil {
		fields = append(fields, sms.FieldCreatedAt)
	}
	if m.updated_at != nil {
		fields = append(fields, sms.FieldUpdatedAt)
	}
	if m.number != nil {
		fields = append(fields, sms.FieldNumber)
	}
	if m.data != nil {
		fields = append(fields, sms.FieldData)
	}
	if m.text != nil {
		fields = append(fields, sms.FieldText)
	}
	if m.discharge_timestamp != nil {
		fields = append(fields, sms.FieldDischargeTimestamp)
	}
	if m.delivery_state != nil {
		fields = append(fields, sms.FieldDeliveryState)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *SMSMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case sms.FieldCreatedAt:
		return m.CreatedAt()
	case sms.FieldUpdatedAt:
		return m.UpdatedAt()
	case sms.FieldNumber:
		return m.Number()
	case sms.FieldData:
		return m.Data()
	case sms.FieldText:
		return m.Text()
	case sms.FieldDischargeTimestamp:
		return m.DischargeTimestamp()
	case sms.FieldDeliveryState:
		return m.DeliveryState()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *SMSMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case sms.FieldCreatedAt:
		return m.OldCreatedAt(ctx)
	case sms.FieldUpdatedAt:
		return m.OldUpdatedAt(ctx)
	case sms.FieldNumber:
		return m.OldNumber(ctx)
	case sms.FieldData:
		return m.OldData(ctx)
	case sms.FieldText:
		return m.OldText(ctx)
	case sms.FieldDischargeTimestamp:
		return m.OldDischargeTimestamp(ctx)
	case sms.FieldDeliveryState:
		return m.OldDeliveryState(ctx)
	}
	return nil, fmt.Errorf("unknown SMS field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *SMSMutation) SetField(name string, value ent.Value) error {
	switch name {
	case sms.FieldCreatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreatedAt(v)
		return nil
	case sms.FieldUpdatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdatedAt(v)
		return nil
	case sms.FieldNumber:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetNumber(v)
		return nil
	case sms.FieldData:
		v, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetData(v)
		return nil
	case sms.FieldText:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetText(v)
		return nil
	case sms.FieldDischargeTimestamp:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDischargeTimestamp(v)
		return nil
	case sms.FieldDeliveryState:
		v, ok := value.(sms.DeliveryState)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDeliveryState(v)
		return nil
	}
	return fmt.Errorf("unknown SMS field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *SMSMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *SMSMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *SMSMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown SMS numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *SMSMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *SMSMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *SMSMutation) ClearField(name string) error {
	return fmt.Errorf("unknown SMS nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *SMSMutation) ResetField(name string) error {
	switch name {
	case sms.FieldCreatedAt:
		m.ResetCreatedAt()
		return nil
	case sms.FieldUpdatedAt:
		m.ResetUpdatedAt()
		return nil
	case sms.FieldNumber:
		m.ResetNumber()
		return nil
	case sms.FieldData:
		m.ResetData()
		return nil
	case sms.FieldText:
		m.ResetText()
		return nil
	case sms.FieldDischargeTimestamp:
		m.ResetDischargeTimestamp()
		return nil
	case sms.FieldDeliveryState:
		m.ResetDeliveryState()
		return nil
	}
	return fmt.Errorf("unknown SMS field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *SMSMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *SMSMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *SMSMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *SMSMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *SMSMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *SMSMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *SMSMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown SMS unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *SMSMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown SMS edge %s", name)
}