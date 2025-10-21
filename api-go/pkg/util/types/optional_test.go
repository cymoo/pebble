package types

import (
	"encoding/json"
	"testing"
)

type UpdateUserRequest struct {
	Name  Optional[string] `json:"name"`
	Email Optional[string] `json:"email"`
	Age   Optional[int]    `json:"age"`
	Bio   Optional[string] `json:"bio"`
}

type User struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Age   int     `json:"age"`
	Bio   *string `json:"bio"`
}

func (u *User) ApplyUpdate(req UpdateUserRequest) {
	if req.Name.IsPresent() {
		u.Name = req.Name.OrDefault("")
	}
	if req.Email.IsPresent() {
		u.Email = req.Email.OrDefault("")
	}
	if req.Age.IsPresent() {
		u.Age = req.Age.OrDefault(0)
	}
	if req.Bio.IsPresent() {
		u.Bio = req.Bio.Ptr()
	}
}

// TestOptionalCreation tests creating Optional values
func TestOptionalCreation(t *testing.T) {
	t.Run("Some creates present value", func(t *testing.T) {
		opt := Some(42)
		if !opt.IsPresent() {
			t.Error("Some() should create present value")
		}
		if opt.IsNull() {
			t.Error("Some() should not create null value")
		}
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("Some(42).Get() = %v, %v; want 42, true", v, ok)
		}
	})

	t.Run("Null creates present null value", func(t *testing.T) {
		opt := Null[string]()
		if !opt.IsPresent() {
			t.Error("Null() should create present value")
		}
		if !opt.IsNull() {
			t.Error("Null() should create null value")
		}
		if _, ok := opt.Get(); ok {
			t.Error("Null().Get() should return false")
		}
	})

	t.Run("Zero value is absent", func(t *testing.T) {
		var opt Optional[int]
		if !opt.IsAbsent() {
			t.Error("Zero value should be absent")
		}
		if opt.IsPresent() {
			t.Error("Zero value should not be present")
		}
		if opt.IsNull() {
			t.Error("Zero value should not be null")
		}
	})
}

// TestOptionalGet tests value retrieval methods
func TestOptionalGet(t *testing.T) {
	t.Run("Get on present value", func(t *testing.T) {
		opt := Some("hello")
		v, ok := opt.Get()
		if !ok {
			t.Error("Get() should return true for present value")
		}
		if v != "hello" {
			t.Errorf("Get() = %q; want %q", v, "hello")
		}
	})

	t.Run("Get on absent value", func(t *testing.T) {
		var opt Optional[string]
		v, ok := opt.Get()
		if ok {
			t.Error("Get() should return false for absent value")
		}
		if v != "" {
			t.Errorf("Get() should return zero value, got %q", v)
		}
	})

	t.Run("Get on null value", func(t *testing.T) {
		opt := Null[int]()
		v, ok := opt.Get()
		if ok {
			t.Error("Get() should return false for null value")
		}
		if v != 0 {
			t.Errorf("Get() should return zero value, got %d", v)
		}
	})

	t.Run("MustGet on present value", func(t *testing.T) {
		opt := Some(100)
		v := opt.MustGet()
		if v != 100 {
			t.Errorf("MustGet() = %d; want 100", v)
		}
	})

	t.Run("MustGet panics on absent", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet() should panic on absent value")
			}
		}()
		var opt Optional[int]
		opt.MustGet()
	})

	t.Run("MustGet panics on null", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet() should panic on null value")
			}
		}()
		opt := Null[int]()
		opt.MustGet()
	})
}

// TestOptionalOrDefault tests default value retrieval
func TestOptionalOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		opt      Optional[string]
		defVal   string
		expected string
	}{
		{"present value", Some("value"), "default", "value"},
		{"absent value", Optional[string]{}, "default", "default"},
		{"null value", Null[string](), "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opt.OrDefault(tt.defVal)
			if result != tt.expected {
				t.Errorf("OrDefault() = %q; want %q", result, tt.expected)
			}
		})
	}
}

// TestOptionalPtr tests pointer retrieval
func TestOptionalPtr(t *testing.T) {
	t.Run("Ptr on present value", func(t *testing.T) {
		opt := Some(42)
		ptr := opt.Ptr()
		if ptr == nil {
			t.Error("Ptr() should not return nil for present value")
		}
		if *ptr != 42 {
			t.Errorf("*Ptr() = %d; want 42", *ptr)
		}
	})

	t.Run("Ptr on absent value", func(t *testing.T) {
		var opt Optional[int]
		ptr := opt.Ptr()
		if ptr != nil {
			t.Error("Ptr() should return nil for absent value")
		}
	})

	t.Run("Ptr on null value", func(t *testing.T) {
		opt := Null[int]()
		ptr := opt.Ptr()
		if ptr != nil {
			t.Error("Ptr() should return nil for null value")
		}
	})
}

// TestOptionalIfPresent tests conditional execution
func TestOptionalIfPresent(t *testing.T) {
	t.Run("IfPresent executes on present value", func(t *testing.T) {
		opt := Some(10)
		executed := false
		opt.IfPresent(func(v int) {
			executed = true
			if v != 10 {
				t.Errorf("IfPresent received %d; want 10", v)
			}
		})
		if !executed {
			t.Error("IfPresent should execute for present value")
		}
	})

	t.Run("IfPresent does not execute on absent", func(t *testing.T) {
		var opt Optional[int]
		executed := false
		opt.IfPresent(func(v int) {
			executed = true
		})
		if executed {
			t.Error("IfPresent should not execute for absent value")
		}
	})

	t.Run("IfPresent does not execute on null", func(t *testing.T) {
		opt := Null[int]()
		executed := false
		opt.IfPresent(func(v int) {
			executed = true
		})
		if executed {
			t.Error("IfPresent should not execute for null value")
		}
	})
}

// TestOptionalUnmarshalJSON tests JSON deserialization
func TestOptionalUnmarshalJSON(t *testing.T) {
	type TestStruct struct {
		Name  Optional[string] `json:"name,omitempty"`
		Age   Optional[int]    `json:"age,omitempty"`
		Email Optional[string] `json:"email,omitempty"`
	}

	t.Run("unmarshal present values", func(t *testing.T) {
		jsonData := `{"name":"Alice","age":30}`
		var ts TestStruct
		if err := json.Unmarshal([]byte(jsonData), &ts); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !ts.Name.IsPresent() || ts.Name.OrDefault("") != "Alice" {
			t.Error("Name should be present with value 'Alice'")
		}
		if !ts.Age.IsPresent() || ts.Age.OrDefault(0) != 30 {
			t.Error("Age should be present with value 30")
		}
		if ts.Email.IsPresent() {
			t.Error("Email should be absent")
		}
	})

	t.Run("unmarshal null values", func(t *testing.T) {
		jsonData := `{"name":"Bob","age":null}`
		var ts TestStruct
		if err := json.Unmarshal([]byte(jsonData), &ts); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !ts.Name.IsPresent() {
			t.Error("Name should be present")
		}
		if !ts.Age.IsPresent() {
			t.Error("Age should be present")
		}
		if !ts.Age.IsNull() {
			t.Error("Age should be null")
		}
		if ts.Email.IsPresent() {
			t.Error("Email should be absent")
		}
	})

	t.Run("unmarshal empty object", func(t *testing.T) {
		jsonData := `{}`
		var ts TestStruct
		if err := json.Unmarshal([]byte(jsonData), &ts); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if ts.Name.IsPresent() {
			t.Error("Name should be absent")
		}
		if ts.Age.IsPresent() {
			t.Error("Age should be absent")
		}
		if ts.Email.IsPresent() {
			t.Error("Email should be absent")
		}
	})

	t.Run("unmarshal mixed states", func(t *testing.T) {
		jsonData := `{"name":"Charlie","age":null,"email":"charlie@example.com"}`
		var ts TestStruct
		if err := json.Unmarshal([]byte(jsonData), &ts); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// name: present with value
		if !ts.Name.IsPresent() || ts.Name.IsNull() {
			t.Error("Name should be present and not null")
		}
		if v, _ := ts.Name.Get(); v != "Charlie" {
			t.Errorf("Name = %q; want 'Charlie'", v)
		}

		// age: present but null
		if !ts.Age.IsPresent() || !ts.Age.IsNull() {
			t.Error("Age should be present and null")
		}

		// email: present with value
		if !ts.Email.IsPresent() || ts.Email.IsNull() {
			t.Error("Email should be present and not null")
		}
	})
}

// TestOptionalMarshalJSON tests JSON serialization
func TestOptionalMarshalJSON(t *testing.T) {
	type TestStruct struct {
		Name Optional[string] `json:"name,omitempty"`
		Age  Optional[int]    `json:"age,omitempty"`
	}

	t.Run("marshal present value", func(t *testing.T) {
		ts := TestStruct{
			Name: Some("Alice"),
			Age:  Some(30),
		}
		data, err := json.Marshal(ts)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal result failed: %v", err)
		}

		if result["name"] != "Alice" {
			t.Errorf("name = %v; want 'Alice'", result["name"])
		}
		if result["age"] != float64(30) {
			t.Errorf("age = %v; want 30", result["age"])
		}
	})

	t.Run("marshal null value", func(t *testing.T) {
		ts := TestStruct{
			Name: Null[string](),
			Age:  Some(25),
		}
		data, err := json.Marshal(ts)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal result failed: %v", err)
		}

		if result["name"] != nil {
			t.Errorf("name should be null, got %v", result["name"])
		}
	})
}

// TestUserUpdateScenario tests real-world update scenario
func TestUserUpdateScenario(t *testing.T) {
	bio := "Original bio"
	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
		Bio:   &bio,
	}

	t.Run("partial update - only email", func(t *testing.T) {
		jsonData := `{"email":"newemail@example.com"}`
		var req UpdateUserRequest
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		originalName := user.Name
		originalAge := user.Age
		originalBio := user.Bio

		user.ApplyUpdate(req)

		if user.Email != "newemail@example.com" {
			t.Errorf("Email not updated: got %q", user.Email)
		}
		if user.Name != originalName {
			t.Error("Name should not change")
		}
		if user.Age != originalAge {
			t.Error("Age should not change")
		}
		if user.Bio != originalBio {
			t.Error("Bio should not change")
		}
	})

	t.Run("set bio to null", func(t *testing.T) {
		jsonData := `{"bio":null}`
		var req UpdateUserRequest
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		user.ApplyUpdate(req)

		if user.Bio != nil {
			t.Error("Bio should be nil after null update")
		}
	})

	t.Run("update multiple fields", func(t *testing.T) {
		jsonData := `{"name":"Jane Doe","age":25,"email":"jane@example.com"}`
		var req UpdateUserRequest
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		user.ApplyUpdate(req)

		if user.Name != "Jane Doe" {
			t.Errorf("Name = %q; want 'Jane Doe'", user.Name)
		}
		if user.Age != 25 {
			t.Errorf("Age = %d; want 25", user.Age)
		}
		if user.Email != "jane@example.com" {
			t.Errorf("Email = %q; want 'jane@example.com'", user.Email)
		}
	})
}

// TestOptionalWithDifferentTypes tests Optional with various types
func TestOptionalWithDifferentTypes(t *testing.T) {
	t.Run("Optional[bool]", func(t *testing.T) {
		opt := Some(true)
		if v, ok := opt.Get(); !ok || !v {
			t.Error("Optional[bool] failed")
		}
	})

	t.Run("Optional[float64]", func(t *testing.T) {
		opt := Some(3.14)
		if v, ok := opt.Get(); !ok || v != 3.14 {
			t.Error("Optional[float64] failed")
		}
	})

	t.Run("Optional[struct]", func(t *testing.T) {
		type Point struct {
			X, Y int
		}
		opt := Some(Point{X: 1, Y: 2})
		if v, ok := opt.Get(); !ok || v.X != 1 || v.Y != 2 {
			t.Error("Optional[struct] failed")
		}
	})

	t.Run("Optional[slice]", func(t *testing.T) {
		opt := Some([]int{1, 2, 3})
		if v, ok := opt.Get(); !ok || len(v) != 3 {
			t.Error("Optional[slice] failed")
		}
	})
}
