package verifications

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Id           int    `json:"id"`
	Confirmation bool   `json:"confirmation"`
	Massage      string `json:"massage"`
}
