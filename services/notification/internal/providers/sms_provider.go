package providers

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	SendSMS(to string, content string) error
}

