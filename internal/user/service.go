package user

type Service interface {
	GetUserByID(id int) *User
}

type mockService struct{}
type mysqlService struct{}

func NewMockService() Service {
	return &mockService{}
}

func NewMySQLService() Service {
	return &mysqlService{}
}

func (s *mockService) GetUserByID(id int) *User {
	return GetUserByID(id)
}

func (s *mysqlService) GetUserByID(id int) *User {
	return nil
}
