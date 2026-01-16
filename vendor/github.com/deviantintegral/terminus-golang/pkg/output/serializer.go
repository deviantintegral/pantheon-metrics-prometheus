// Package output provides formatting utilities for CLI output.
package output

// Serializer is an interface that models can implement to control
// their CSV and table output formatting. This provides PHP Terminus-like
// behavior where models explicitly define field order and names.
type Serializer interface {
	// Serialize returns an ordered list of fields for output.
	// The order of fields in the returned slice determines the column order
	// in CSV and table outputs.
	Serialize() []SerializedField
}

// SerializedField represents a single field in serialized output.
type SerializedField struct {
	// Name is the field name as it should appear in CSV headers and table columns
	Name string
	// Value is the field value
	Value interface{}
}

// DefaultFielder is an optional interface that models can implement
// to specify which fields should be shown by default (similar to PHP's @default-fields).
type DefaultFielder interface {
	// DefaultFields returns the list of field names to show by default
	DefaultFields() []string
}
