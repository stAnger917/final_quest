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

type OrderInfo struct {
	ID     int
	UserID int
	Number string
}

type UserBalanceInfo struct {
	Current  float32 `json:"current"`
	Withdraw float32 `json:"withdraw"`
}

type UserBalance struct {
	UserID   int     `json:"user_id"`
	Current  float32 `json:"current"`
	Withdraw float32 `json:"withdraw"`
}
