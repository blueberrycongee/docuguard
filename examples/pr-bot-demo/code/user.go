package demo

import "errors"

// User represents a user in the system.
type User struct {
	ID    int
	Name  string
	Email string
}

// CreateUser creates a new user with the given name.
// Email will be auto-generated based on the name.
// Returns the created user and an error if validation fails.
func CreateUser(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	user := &User{
		ID:    generateID(),
		Name:  name,
		Email: generateEmail(name),
	}
	return user, nil
}

func generateEmail(name string) string {
	// Auto-generate email from name
	return name + "@example.com"
}

// GetUserByID retrieves a user by their ID.
func GetUserByID(id int) (*User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}
	// Simulate database lookup
	return &User{ID: id, Name: "John Doe", Email: "john@example.com"}, nil
}

// UpdateUser updates an existing user's information.
func UpdateUser(id int, name, email string) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}
	// Simulate update operation
	return nil
}

// DeleteUser removes a user from the system.
func DeleteUser(id int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}
	// Simulate delete operation
	return nil
}

func generateID() int {
	// Simulate ID generation
	return 12345
}
