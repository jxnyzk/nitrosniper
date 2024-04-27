package auth

type Res struct {
	Suc 	bool `json:"success"`
	Hash 	string `json:"hash"`
	User 	string `json:"user"`
}