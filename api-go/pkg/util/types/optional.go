package types

import (
	"encoding/json"
)

// Optional represents a value that may be absent, null, or present.
// The zero value represents an absent field.
type Optional[T any] struct {
	value   *T
	present bool
}

// Some creates an Optional with a non-nil value
func Some[T any](value T) Optional[T] {
	return Optional[T]{value: &value, present: true}
}

// Null creates an Optional that is present but has a null value
func Null[T any]() Optional[T] {
	return Optional[T]{value: nil, present: true}
}

// IsPresent returns true if the field exists (including null)
func (o Optional[T]) IsPresent() bool {
	return o.present
}

// IsAbsent returns true if the field is missing
func (o Optional[T]) IsAbsent() bool {
	return !o.present
}

// IsNull returns true if the field is present but null
func (o Optional[T]) IsNull() bool {
	return o.present && o.value == nil
}

// Get returns the value and true if present and non-null, otherwise zero value and false
func (o Optional[T]) Get() (T, bool) {
	var zero T
	if !o.present || o.value == nil {
		return zero, false
	}
	return *o.value, true
}

// MustGet returns the value or panics if absent or null
func (o Optional[T]) MustGet() T {
	if v, ok := o.Get(); ok {
		return v
	}
	panic("Optional: value is absent or null")
}

// OrDefault returns the value or the default if absent or null
func (o Optional[T]) OrDefault(defaultValue T) T {
	if v, ok := o.Get(); ok {
		return v
	}
	return defaultValue
}

// Ptr returns a pointer to the value (nil for absent or null)
func (o Optional[T]) Ptr() *T {
	if !o.present {
		return nil
	}
	return o.value
}

// IfPresent executes the function if value is present and non-null
func (o Optional[T]) IfPresent(fn func(T)) {
	if v, ok := o.Get(); ok {
		fn(v)
	}
}

// UnmarshalJSON is called only when the field exists in JSON
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.present = true // Field exists in JSON

	if string(data) == "null" {
		o.value = nil
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.value = &value
	return nil
}

// MarshalJSON serializes the Optional
// Note: For proper omitempty behavior with absent fields,
// consider using *Optional in struct fields
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.present {
		// This will still serialize as null, not omit the field
		// To truly omit, use *Optional[T] with omitempty
		return []byte("null"), nil
	}
	if o.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*o.value)
}
