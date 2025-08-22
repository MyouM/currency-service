package repository

var (
	secret = []byte("Some_Secret_Text")
	//tokens = make(map[string]struct{})
	users = make(map[string]string)
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u User) Add() {
	users[u.Login] = u.Password
}

func (u User) Exist() bool {
	_, ok := users[u.Login]
	return ok
}

//func FindToken(token string) bool {
//	_, ok := tokens[token]
//	return ok
//}

//func AddToken(token string) {
//	tokens[token] = struct{}{}
//}

func GetSecret() []byte {
	return secret
}
