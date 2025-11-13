package example

// User represents a user in the system
type User struct {
	ID    int
	Name  string
	Email string
}

// UserService provides user operations
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
}

// DatabaseUserService implements UserService
type DatabaseUserService struct {
	DB any // placeholder for database connection
}

func (s *DatabaseUserService) GetUser(id int) (*User, error) {
	return nil, nil
}

func (s *DatabaseUserService) CreateUser(user *User) error {
	return nil
}
