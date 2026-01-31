// Package i18n provides simple localization utilities.
package i18n

import (
	"fmt"
	"strings"
	"sync"
)

// Localizer resolves localized strings for a locale.
type Localizer interface {
	Locale() string
	T(key string) string
	Tf(key string, args ...any) string
}

// Bundle stores localized messages by locale.
type Bundle struct {
	mu       sync.RWMutex
	fallback string
	messages map[string]map[string]string
}

// NewBundle creates a bundle with a fallback locale.
func NewBundle(fallback string) *Bundle {
	return &Bundle{
		fallback: strings.TrimSpace(fallback),
		messages: map[string]map[string]string{},
	}
}

// AddMessages registers a message map for the locale.
func (b *Bundle) AddMessages(locale string, messages map[string]string) {
	if b == nil {
		return
	}
	locale = strings.TrimSpace(locale)
	if locale == "" || len(messages) == 0 {
		return
	}
	b.mu.Lock()
	if b.messages == nil {
		b.messages = map[string]map[string]string{}
	}
	copyMap := make(map[string]string, len(messages))
	for key, value := range messages {
		copyMap[key] = value
	}
	b.messages[locale] = copyMap
	b.mu.Unlock()
}

// Localizer returns a locale-bound resolver.
func (b *Bundle) Localizer(locale string) *BundleLocalizer {
	if b == nil {
		return &BundleLocalizer{}
	}
	return &BundleLocalizer{bundle: b, locale: strings.TrimSpace(locale)}
}

// Translate resolves a key for the locale.
func (b *Bundle) Translate(locale, key string) string {
	return b.translate(locale, key, nil)
}

// Translatef resolves a key with formatting args.
func (b *Bundle) Translatef(locale, key string, args ...any) string {
	return b.translate(locale, key, args)
}

func (b *Bundle) translate(locale, key string, args []any) string {
	if b == nil {
		return formatFallback(key, args...)
	}
	locale = strings.TrimSpace(locale)
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	b.mu.RLock()
	msg := lookupMessage(b.messages, locale, key)
	if msg == "" && b.fallback != "" {
		msg = lookupMessage(b.messages, b.fallback, key)
	}
	b.mu.RUnlock()

	if msg == "" {
		return formatFallback(key, args...)
	}
	return formatMessage(msg, args...)
}

// BundleLocalizer binds a bundle to a locale.
type BundleLocalizer struct {
	bundle *Bundle
	locale string
}

// Locale returns the locale identifier.
func (l *BundleLocalizer) Locale() string {
	if l == nil {
		return ""
	}
	return l.locale
}

// T translates a key with optional format args.
func (l *BundleLocalizer) T(key string) string {
	if l == nil || l.bundle == nil {
		return formatFallback(key)
	}
	return l.bundle.Translate(l.locale, key)
}

// Tf translates a key with format args.
func (l *BundleLocalizer) Tf(key string, args ...any) string {
	if l == nil || l.bundle == nil {
		return formatFallback(key, args...)
	}
	return l.bundle.Translatef(l.locale, key, args...)
}

// MapLocalizer is a lightweight localizer backed by a map.
type MapLocalizer struct {
	LocaleCode string
	Messages   map[string]string
	Fallback   map[string]string
}

// Locale returns the locale identifier.
func (m MapLocalizer) Locale() string { return m.LocaleCode }

// T resolves a key from the map.
func (m MapLocalizer) T(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if msg, ok := m.Messages[key]; ok {
		return msg
	}
	if msg, ok := m.Fallback[key]; ok {
		return msg
	}
	return key
}

// Tf resolves a key with format args.
func (m MapLocalizer) Tf(key string, args ...any) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if msg, ok := m.Messages[key]; ok {
		return formatMessage(msg, args...)
	}
	if msg, ok := m.Fallback[key]; ok {
		return formatMessage(msg, args...)
	}
	return formatFallback(key, args...)
}

func lookupMessage(messages map[string]map[string]string, locale, key string) string {
	if messages == nil {
		return ""
	}
	if locale == "" {
		return ""
	}
	localeMap, ok := messages[locale]
	if !ok || localeMap == nil {
		return ""
	}
	return localeMap[key]
}

func formatMessage(msg string, args ...any) string {
	return applyArgs(msg, args...)
}

func formatFallback(key string, args ...any) string {
	return applyArgs(key, args...)
}

func applyArgs(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	out := msg
	for i, arg := range args {
		replaced := false
		token := fmt.Sprintf("{%d}", i)
		if strings.Contains(out, token) {
			out = strings.ReplaceAll(out, token, fmt.Sprint(arg))
			replaced = true
		}
		if !replaced && strings.Contains(out, "{}") {
			out = strings.Replace(out, "{}", fmt.Sprint(arg), 1)
		}
	}
	return out
}

var _ Localizer = (*BundleLocalizer)(nil)
var _ Localizer = (*MapLocalizer)(nil)
