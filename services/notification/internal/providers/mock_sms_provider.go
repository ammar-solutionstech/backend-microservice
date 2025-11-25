package providers

import "log"

// MockSMSProvider is a mock implementation of SMSProvider for testing
type MockSMSProvider struct{}

func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{}
}

func (p *MockSMSProvider) SendSMS(to string, content string) error {
	log.Printf("[MOCK SMS] Sending SMS to %s: %s", to, content)
	// In a real implementation, this would call an SMS service like Twilio
	return nil
}

