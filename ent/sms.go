// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/lixin9311/sms-bot/ent/sms"
)

// SMS is the model entity for the SMS schema.
type SMS struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// ModemID holds the value of the "modem_id" field.
	ModemID *string `json:"modem_id,omitempty"`
	// Number holds the value of the "number" field.
	Number string `json:"number,omitempty"`
	// Data holds the value of the "data" field.
	Data []byte `json:"data,omitempty"`
	// Text holds the value of the "text" field.
	Text string `json:"text,omitempty"`
	// DischargeTimestamp holds the value of the "discharge_timestamp" field.
	DischargeTimestamp time.Time `json:"discharge_timestamp,omitempty"`
	// DeliveryState holds the value of the "delivery_state" field.
	DeliveryState sms.DeliveryState `json:"delivery_state,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SMS) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case sms.FieldData:
			values[i] = new([]byte)
		case sms.FieldID:
			values[i] = new(sql.NullInt64)
		case sms.FieldModemID, sms.FieldNumber, sms.FieldText, sms.FieldDeliveryState:
			values[i] = new(sql.NullString)
		case sms.FieldCreatedAt, sms.FieldUpdatedAt, sms.FieldDischargeTimestamp:
			values[i] = new(sql.NullTime)
		default:
			return nil, fmt.Errorf("unexpected column %q for type SMS", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SMS fields.
func (s *SMS) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case sms.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			s.ID = int(value.Int64)
		case sms.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				s.CreatedAt = value.Time
			}
		case sms.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				s.UpdatedAt = value.Time
			}
		case sms.FieldModemID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field modem_id", values[i])
			} else if value.Valid {
				s.ModemID = new(string)
				*s.ModemID = value.String
			}
		case sms.FieldNumber:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field number", values[i])
			} else if value.Valid {
				s.Number = value.String
			}
		case sms.FieldData:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field data", values[i])
			} else if value != nil {
				s.Data = *value
			}
		case sms.FieldText:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field text", values[i])
			} else if value.Valid {
				s.Text = value.String
			}
		case sms.FieldDischargeTimestamp:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field discharge_timestamp", values[i])
			} else if value.Valid {
				s.DischargeTimestamp = value.Time
			}
		case sms.FieldDeliveryState:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field delivery_state", values[i])
			} else if value.Valid {
				s.DeliveryState = sms.DeliveryState(value.String)
			}
		}
	}
	return nil
}

// Update returns a builder for updating this SMS.
// Note that you need to call SMS.Unwrap() before calling this method if this SMS
// was returned from a transaction, and the transaction was committed or rolled back.
func (s *SMS) Update() *SMSUpdateOne {
	return (&SMSClient{config: s.config}).UpdateOne(s)
}

// Unwrap unwraps the SMS entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (s *SMS) Unwrap() *SMS {
	tx, ok := s.config.driver.(*txDriver)
	if !ok {
		panic("ent: SMS is not a transactional entity")
	}
	s.config.driver = tx.drv
	return s
}

// String implements the fmt.Stringer.
func (s *SMS) String() string {
	var builder strings.Builder
	builder.WriteString("SMS(")
	builder.WriteString(fmt.Sprintf("id=%v", s.ID))
	builder.WriteString(", created_at=")
	builder.WriteString(s.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", updated_at=")
	builder.WriteString(s.UpdatedAt.Format(time.ANSIC))
	if v := s.ModemID; v != nil {
		builder.WriteString(", modem_id=")
		builder.WriteString(*v)
	}
	builder.WriteString(", number=")
	builder.WriteString(s.Number)
	builder.WriteString(", data=")
	builder.WriteString(fmt.Sprintf("%v", s.Data))
	builder.WriteString(", text=")
	builder.WriteString(s.Text)
	builder.WriteString(", discharge_timestamp=")
	builder.WriteString(s.DischargeTimestamp.Format(time.ANSIC))
	builder.WriteString(", delivery_state=")
	builder.WriteString(fmt.Sprintf("%v", s.DeliveryState))
	builder.WriteByte(')')
	return builder.String()
}

// SMSs is a parsable slice of SMS.
type SMSs []*SMS

func (s SMSs) config(cfg config) {
	for _i := range s {
		s[_i].config = cfg
	}
}
