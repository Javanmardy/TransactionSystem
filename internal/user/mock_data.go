package user

var mockUsers = []User{
	{ID: 1, Username: "amirhossein", Email: "amirhossein@example.com"},
	{ID: 2, Username: "armin", Email: "armin@example.com"},
	{ID: 3, Username: "sara", Email: "sara@example.com"},
}

func GetUserByID(id int) *User {
	for _, u := range mockUsers {
		if u.ID == id {
			return &u
		}
	}
	return nil
}
