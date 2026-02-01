package i18n

import (
	"strings"
	"sync"
)

// PluralForm identifies a plural category.
type PluralForm string

const (
	PluralZero  PluralForm = "zero"
	PluralOne   PluralForm = "one"
	PluralTwo   PluralForm = "two"
	PluralFew   PluralForm = "few"
	PluralMany  PluralForm = "many"
	PluralOther PluralForm = "other"
)

// PluralRule maps a count to a plural form.
type PluralRule func(n int) PluralForm

var (
	pluralMu    sync.RWMutex
	pluralRules = map[string]PluralRule{
		"en": pluralRuleEnglish,
		"fr": pluralRuleFrench,
		"ru": pluralRuleRussian,
		"ar": pluralRuleArabic,
		"ja": pluralRuleOther,
		"zh": pluralRuleOther,
		"ko": pluralRuleOther,
		"tr": pluralRuleOther,
	}

	directionMu        sync.RWMutex
	directionOverrides = map[string]Direction{}
	rtlLocales         = map[string]bool{
		"ar": true,
		"fa": true,
		"he": true,
		"ur": true,
		"dv": true,
	}
)

// Direction represents text direction for a locale.
type Direction int

const (
	DirectionLTR Direction = iota
	DirectionRTL
)

// RegisterPluralRule overrides the plural rule for a locale.
func RegisterPluralRule(locale string, rule PluralRule) {
	locale = normalizeLocale(locale)
	if locale == "" || rule == nil {
		return
	}
	pluralMu.Lock()
	pluralRules[locale] = rule
	pluralMu.Unlock()
}

// PluralFormForLocale resolves the plural form for a locale and count.
func PluralFormForLocale(locale string, n int) PluralForm {
	rule := pluralRuleForLocale(locale)
	return rule(n)
}

// TranslatePlural resolves a pluralized key for the locale.
// It looks for "{key}.{form}", falling back to "{key}.other" and "{key}".
// The count is passed as the first format argument.
func (b *Bundle) TranslatePlural(locale, key string, count int, args ...any) string {
	return b.translatePlural(locale, key, count, args)
}

// Tn translates a pluralized key with the localizer locale.
func (l *BundleLocalizer) Tn(key string, count int, args ...any) string {
	if l == nil || l.bundle == nil {
		return formatFallback(key, append([]any{count}, args...)...)
	}
	return l.bundle.TranslatePlural(l.locale, key, count, args...)
}

// Tn translates a pluralized key using the map localizer.
func (m MapLocalizer) Tn(key string, count int, args ...any) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	form := PluralFormForLocale(m.LocaleCode, count)
	msg, ok := m.Messages[key+"."+string(form)]
	if !ok {
		msg, ok = m.Messages[key+"."+string(PluralOther)]
	}
	if !ok {
		msg, ok = m.Messages[key]
	}
	if !ok {
		if fallback, ok := m.Fallback[key+"."+string(form)]; ok {
			msg = fallback
		} else if fallback, ok := m.Fallback[key+"."+string(PluralOther)]; ok {
			msg = fallback
		} else if fallback, ok := m.Fallback[key]; ok {
			msg = fallback
		} else {
			return formatFallback(key, append([]any{count}, args...)...)
		}
	}
	return formatMessage(msg, append([]any{count}, args...)...)
}

// DirectionForLocale returns text direction for a locale.
func DirectionForLocale(locale string) Direction {
	locale = normalizeLocale(locale)
	if locale == "" {
		return DirectionLTR
	}
	directionMu.RLock()
	if dir, ok := directionOverrides[locale]; ok {
		directionMu.RUnlock()
		return dir
	}
	directionMu.RUnlock()
	if rtlLocales[locale] {
		return DirectionRTL
	}
	return DirectionLTR
}

// IsRTL reports whether the locale is right-to-left.
func IsRTL(locale string) bool {
	return DirectionForLocale(locale) == DirectionRTL
}

// RegisterDirection overrides the direction for a locale.
func RegisterDirection(locale string, dir Direction) {
	locale = normalizeLocale(locale)
	if locale == "" {
		return
	}
	directionMu.Lock()
	directionOverrides[locale] = dir
	directionMu.Unlock()
}

func (b *Bundle) translatePlural(locale, key string, count int, args []any) string {
	if b == nil {
		return formatFallback(key, append([]any{count}, args...)...)
	}
	locale = strings.TrimSpace(locale)
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	form := PluralFormForLocale(locale, count)
	pluralKey := key + "." + string(form)

	b.mu.RLock()
	msg := lookupMessage(b.messages, locale, pluralKey)
	if msg == "" {
		msg = lookupMessage(b.messages, locale, key+"."+string(PluralOther))
	}
	if msg == "" {
		msg = lookupMessage(b.messages, locale, key)
	}
	if msg == "" && b.fallback != "" {
		msg = lookupMessage(b.messages, b.fallback, pluralKey)
		if msg == "" {
			msg = lookupMessage(b.messages, b.fallback, key+"."+string(PluralOther))
		}
		if msg == "" {
			msg = lookupMessage(b.messages, b.fallback, key)
		}
	}
	b.mu.RUnlock()

	if msg == "" {
		return formatFallback(key, append([]any{count}, args...)...)
	}
	return formatMessage(msg, append([]any{count}, args...)...)
}

func pluralRuleForLocale(locale string) PluralRule {
	locale = normalizeLocale(locale)
	pluralMu.RLock()
	rule, ok := pluralRules[locale]
	pluralMu.RUnlock()
	if ok {
		return rule
	}
	return pluralRuleEnglish
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		return ""
	}
	if idx := strings.IndexAny(locale, "-_"); idx != -1 {
		return locale[:idx]
	}
	return locale
}

func pluralRuleEnglish(n int) PluralForm {
	if n == 1 {
		return PluralOne
	}
	return PluralOther
}

func pluralRuleFrench(n int) PluralForm {
	if n == 0 || n == 1 {
		return PluralOne
	}
	return PluralOther
}

func pluralRuleRussian(n int) PluralForm {
	mod10 := n % 10
	mod100 := n % 100
	switch {
	case mod10 == 1 && mod100 != 11:
		return PluralOne
	case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
		return PluralFew
	case mod10 == 0 || (mod10 >= 5 && mod10 <= 9) || (mod100 >= 11 && mod100 <= 14):
		return PluralMany
	default:
		return PluralOther
	}
}

func pluralRuleArabic(n int) PluralForm {
	switch {
	case n == 0:
		return PluralZero
	case n == 1:
		return PluralOne
	case n == 2:
		return PluralTwo
	case n%100 >= 3 && n%100 <= 10:
		return PluralFew
	case n%100 >= 11 && n%100 <= 99:
		return PluralMany
	default:
		return PluralOther
	}
}

func pluralRuleOther(int) PluralForm {
	return PluralOther
}
