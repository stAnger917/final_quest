package models

type RegistrationData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserData struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderData struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type SingleOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
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

type OrderOwner struct {
	UserID int
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

type WithdrawData struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

type OrderAccountingInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type OrderNumber struct {
	Number string
}

type Withdrawals struct {
	Data []WithdrawInfo
}

type WithdrawInfo struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
