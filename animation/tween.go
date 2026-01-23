package animation

import "time"

// Tween represents a single property animation.
type Tween struct {
	getValue func() Animatable
	setValue func(Animatable)

	startValue Animatable
	endValue   Animatable
	startTime  time.Time
	duration   time.Duration
	delay      time.Duration
	easing     EasingFunc

	onUpdate   func(value Animatable)
	onComplete func()

	started   bool
	completed bool
	paused    bool
}

// TweenConfig configures a tween.
type TweenConfig struct {
	Duration   time.Duration
	Delay      time.Duration
	Easing     EasingFunc
	OnUpdate   func(value Animatable)
	OnComplete func()
}

// NewTween creates a tween.
func NewTween(getValue func() Animatable, setValue func(Animatable), endValue Animatable, cfg TweenConfig) *Tween {
	if cfg.Easing == nil {
		cfg.Easing = OutCubic
	}
	if cfg.Duration == 0 {
		cfg.Duration = 300 * time.Millisecond
	}
	return &Tween{
		getValue:   getValue,
		setValue:   setValue,
		endValue:   endValue,
		duration:   cfg.Duration,
		delay:      cfg.Delay,
		easing:     cfg.Easing,
		onUpdate:   cfg.OnUpdate,
		onComplete: cfg.OnComplete,
	}
}

// Start begins the animation.
func (t *Tween) Start() {
	if t == nil {
		return
	}
	t.startTime = time.Now()
	if t.getValue != nil {
		t.startValue = t.getValue()
	}
	t.started = true
	t.completed = false
}

// Update advances the animation, returns true when complete.
func (t *Tween) Update(now time.Time) bool {
	if t == nil || t.completed || t.paused || !t.started {
		return t != nil && t.completed
	}
	elapsed := now.Sub(t.startTime)
	if elapsed < t.delay {
		return false
	}
	if t.duration <= 0 {
		t.setValue(t.endValue)
		t.completed = true
		if t.onUpdate != nil {
			t.onUpdate(t.endValue)
		}
		if t.onComplete != nil {
			t.onComplete()
		}
		return true
	}
	animElapsed := elapsed - t.delay
	progress := float64(animElapsed) / float64(t.duration)
	if progress >= 1.0 {
		progress = 1.0
		t.completed = true
	}
	eased := t.easing(progress)
	if t.startValue == nil && t.getValue != nil {
		t.startValue = t.getValue()
	}
	if t.startValue != nil {
		value := t.startValue.Lerp(t.endValue, eased)
		if t.setValue != nil {
			t.setValue(value)
		}
		if t.onUpdate != nil {
			t.onUpdate(value)
		}
	}
	if t.completed && t.onComplete != nil {
		t.onComplete()
	}
	return t.completed
}

// Pause pauses the animation.
func (t *Tween) Pause() {
	if t == nil {
		return
	}
	t.paused = true
}

// Resume resumes the animation.
func (t *Tween) Resume() {
	if t == nil {
		return
	}
	t.paused = false
}

// Stop stops the animation.
func (t *Tween) Stop() {
	if t == nil {
		return
	}
	t.completed = true
}

// Complete jumps to the end.
func (t *Tween) Complete() {
	if t == nil {
		return
	}
	if t.setValue != nil {
		t.setValue(t.endValue)
	}
	t.completed = true
	if t.onComplete != nil {
		t.onComplete()
	}
}
