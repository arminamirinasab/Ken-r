// Package sms holds SMSProvider adapters. LogProvider is the development
// default that "sends" by logging; the production Kavenegar adapter implements
// the same ports.SMSProvider interface and swaps in via config.
package sms

import (
	"context"
	"log/slog"
)

// LogProvider satisfies ports.SMSProvider by logging the message instead of
// sending it. For LOCAL DEVELOPMENT ONLY — it logs the OTP code in clear text,
// which must never run against real users. Selected when no real SMS provider
// is configured.
type LogProvider struct {
	log *slog.Logger
}

// NewLogProvider constructs a logging SMS provider.
func NewLogProvider(log *slog.Logger) *LogProvider {
	return &LogProvider{log: log}
}

// Send logs the would-be SMS.
func (p *LogProvider) Send(_ context.Context, phone, message string) error {
	p.log.Warn("DEV sms (not actually sent)", "to", phone, "body", message)
	return nil
}
