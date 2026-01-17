package keybind

// Command represents a named action.
type Command struct {
	ID          string
	Title       string
	Description string
	Category    string
	Handler     func(ctx Context)
	Enabled     func(ctx Context) bool
}

// CommandRegistry stores registered commands.
type CommandRegistry struct {
	commands map[string]Command
}

// NewRegistry returns an empty registry.
func NewRegistry() *CommandRegistry {
	return &CommandRegistry{commands: make(map[string]Command)}
}

// Register registers or replaces a command.
func (r *CommandRegistry) Register(cmd Command) {
	if r == nil || cmd.ID == "" {
		return
	}
	if r.commands == nil {
		r.commands = make(map[string]Command)
	}
	r.commands[cmd.ID] = cmd
}

// RegisterAll registers multiple commands.
func (r *CommandRegistry) RegisterAll(cmds ...Command) {
	for _, cmd := range cmds {
		r.Register(cmd)
	}
}

// Get returns a command by ID.
func (r *CommandRegistry) Get(id string) (Command, bool) {
	if r == nil {
		return Command{}, false
	}
	cmd, ok := r.commands[id]
	return cmd, ok
}

// Execute runs a command if registered and enabled.
func (r *CommandRegistry) Execute(id string, ctx Context) bool {
	if r == nil || id == "" {
		return false
	}
	cmd, ok := r.commands[id]
	if !ok || cmd.Handler == nil {
		return false
	}
	if cmd.Enabled != nil && !cmd.Enabled(ctx) {
		return false
	}
	cmd.Handler(ctx)
	return true
}

// List returns registered commands.
func (r *CommandRegistry) List() []Command {
	if r == nil || len(r.commands) == 0 {
		return nil
	}
	out := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		out = append(out, cmd)
	}
	return out
}
