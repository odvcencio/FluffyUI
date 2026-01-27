package gpu

// drawBatch collects draw calls for submission.
type drawBatch struct {
	calls []DrawCall
}

func (b *drawBatch) Reset() {
	if b == nil {
		return
	}
	b.calls = b.calls[:0]
}

func (b *drawBatch) Add(call DrawCall) {
	if b == nil {
		return
	}
	b.calls = append(b.calls, call)
}

func (b *drawBatch) Calls() []DrawCall {
	if b == nil {
		return nil
	}
	return b.calls
}
