# User API Documentation

## Overview

The User API provides functionality for managing users in the system.

## Functions

### CreateUser

Creates a new user in the system.

**Signature:**
```go
func CreateUser(name string) (*User, error)
```

**Parameters:**
- `name` (string): The user's full name

**Returns:**
- `*User`: The created user object
- `error`: Error if validation fails

**Example:**
```go
user, err := CreateUser("John Doe")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created user: %s\n", user.Name)
```

**Note:** The name parameter is required and cannot be empty.

---

### GetUserByID

Retrieves a user by their unique ID.

**Signature:**
```go
func GetUserByID(id int) (*User, error)
```

**Parameters:**
- `id` (int): The user's unique identifier

**Returns:**
- `*User`: The user object if found
- `error`: Error if user not found or invalid ID

**Example:**
```go
user, err := GetUserByID(12345)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
```

---

### UpdateUser

Updates an existing user's information.

**Signature:**
```go
func UpdateUser(id int, name, email string) error
```

**Parameters:**
- `id` (int): The user's unique identifier
- `name` (string): New name for the user
- `email` (string): New email for the user

**Returns:**
- `error`: Error if update fails

**Example:**
```go
err := UpdateUser(12345, "Jane Doe", "jane@example.com")
if err != nil {
    log.Fatal(err)
}
```

---

### DeleteUser

Removes a user from the system.

**Signature:**
```go
func DeleteUser(id int) error
```

**Parameters:**
- `id` (int): The user's unique identifier

**Returns:**
- `error`: Error if deletion fails

**Example:**
```go
err := DeleteUser(12345)
if err != nil {
    log.Fatal(err)
}
```

**Warning:** This operation is permanent and cannot be undone.
