package models

type RegistrationData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserData struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderData struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type SingleOrderData struct {
	Number string
	UserID int
}
