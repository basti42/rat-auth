package models

type Token struct {
	Id       string `gorm:"primaryKey" json:"id"`
	Expires  string `gorm:"not null" json:"expires"`
	UserId   string `gorm:"null" json:"user_id"`
	State    string `gorm:"null" json:"state"`
	Verifier string `gorm:"null" json:"verifier"`
}

type User struct {
	Id     string `gorm:"primaryKey" json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Sub    string `json:"sub"`
	Avatar string `json:"avatar"`
}
