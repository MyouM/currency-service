package repository

var (
	secret = []byte("Some_Secret_Text")
	tokens = make(map[string]struct{})
)

func FindToken(token string) bool {
	_, ok := tokens[token]
	return ok
}

func AddToken(token string) {
	tokens[token] = struct{}{}
}

func GetSecret() []byte {
	return secret
}
