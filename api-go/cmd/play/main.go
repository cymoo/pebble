package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cymoo/pebble/internal/app"
	"github.com/cymoo/pebble/internal/config"
	t "github.com/cymoo/pebble/pkg/util/types"
)

func main1() {
	expirePost(1)
}

// expirePost sets the deleted_at field of a post to 31 days ago, effectively expiring it
// postID: ID of the post to expire
func expirePost(postID int64) {
	cfg := config.Load()

	application := app.New(cfg)

	db := application.GetDB()

	thirtyOneDaysAgo := time.Now().UTC().AddDate(0, 0, -31).UnixMilli()

	_, err := db.Exec(
		"UPDATE posts SET deleted_at = $1 WHERE id = $2",
		thirtyOneDaysAgo,
		postID,
	)
	if err != nil {
		panic(err)
	}
}

// ===== Example Usage =====

type UpdateUserRequest struct {
	Name  t.Optional[string] `json:"name,omitempty"`
	Email t.Optional[string] `json:"email,omitempty"`
	Age   t.Optional[int]    `json:"age,omitempty"`
	Bio   t.Optional[string] `json:"bio,omitempty"`
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

func main() {
	// Test 1: Missing field
	fmt.Println("=== Test 1: Missing field ===")
	json1 := `{"name": "Alice"}`
	var req1 UpdateUserRequest
	json.Unmarshal([]byte(json1), &req1)

	fmt.Printf("name - present: %v, value: %q\n", req1.Name.IsPresent(), req1.Name.OrDefault("N/A"))
	fmt.Printf("email - present: %v (absent: %v)\n", req1.Email.IsPresent(), req1.Email.IsAbsent())
	fmt.Printf("age - present: %v (absent: %v)\n", req1.Age.IsPresent(), req1.Age.IsAbsent())

	// Test 2: Null value
	fmt.Println("\n=== Test 2: Null value ===")
	json2 := `{"name": "Bob", "bio": null}`
	var req2 UpdateUserRequest
	json.Unmarshal([]byte(json2), &req2)

	fmt.Printf("name - present: %v, null: %v, value: %q\n",
		req2.Name.IsPresent(), req2.Name.IsNull(), req2.Name.OrDefault("N/A"))
	fmt.Printf("bio - present: %v, null: %v\n",
		req2.Bio.IsPresent(), req2.Bio.IsNull())
	fmt.Printf("email - present: %v (absent: %v)\n",
		req2.Email.IsPresent(), req2.Email.IsAbsent())

	// Test 3: Apply update
	fmt.Println("\n=== Test 3: Apply update ===")
	bio := "Original bio"
	user := User{
		ID:    1,
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
		Bio:   &bio,
	}
	fmt.Printf("Before: Name=%s, Email=%s, Age=%d, Bio=%v\n",
		user.Name, user.Email, user.Age, *user.Bio)

	json3 := `{"email": "newemail@example.com", "bio": null}`
	var req3 UpdateUserRequest
	json.Unmarshal([]byte(json3), &req3)

	user.ApplyUpdate(req3)
	bioStr := "nil"
	if user.Bio != nil {
		bioStr = *user.Bio
	}
	fmt.Printf("After: Name=%s, Email=%s, Age=%d, Bio=%s\n",
		user.Name, user.Email, user.Age, bioStr)

	// Test 4: Three states comparison
	fmt.Println("\n=== Test 4: Three states comparison ===")
	json4 := `{"name": "Test", "email": null}`
	var req4 UpdateUserRequest
	json.Unmarshal([]byte(json4), &req4)

	fmt.Printf("name  - absent: %v, null: %v, has value: %v\n",
		req4.Name.IsAbsent(), req4.Name.IsNull(), req4.Name.IsPresent() && !req4.Name.IsNull())
	fmt.Printf("email - absent: %v, null: %v, has value: %v\n",
		req4.Email.IsAbsent(), req4.Email.IsNull(), req4.Email.IsPresent() && !req4.Email.IsNull())
	fmt.Printf("bio   - absent: %v, null: %v, has value: %v\n",
		req4.Bio.IsAbsent(), req4.Bio.IsNull(), req4.Bio.IsPresent() && !req4.Bio.IsNull())
}
