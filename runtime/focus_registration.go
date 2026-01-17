package runtime

// FocusRegistrationMode configures how focusables are registered.
type FocusRegistrationMode int

const (
	// FocusRegistrationManual requires apps to register focusables explicitly.
	FocusRegistrationManual FocusRegistrationMode = iota
	// FocusRegistrationAuto scans widget trees on root/layer changes.
	FocusRegistrationAuto
)
