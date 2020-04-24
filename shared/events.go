package shared

type MailMessage struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Content   string `json:"content"`
}
