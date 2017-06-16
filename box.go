package isowrap

import "runtime"

// EnvPair represents an environment variable made of a key and a value.
type EnvPair struct {
	Var   string
	Value string
}

// BoxError represents an error encountered after running the program in the box.
type BoxError int

// RunResult represents the result of running the prograim in the box.
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int

	CPUTime   float64
	WallTime  float64
	MemUsed   uint
	ErrorType BoxError
}

// BoxConfig contains configuration data for the BoxRunner
type BoxConfig struct {
	CPUTime      float64
	WallTime     float64
	MemoryLimit  uint
	StackLimit   uint
	MaxProc      uint
	ShareNetwork bool
	FullEnv      bool
	Env          []EnvPair
}

// Runner is an interface for various program isolating methods
type Runner interface {
	Init() error
	Run(string) (RunResult, error)
	Cleanup() error
}

const (
	// NoError means that no error has been returned by the box runner
	NoError BoxError = iota

	// RunTimeError means that an error was rised at run time. Probably non-zero status.
	RunTimeError = iota

	// KilledBySignal means that the program was killed after getting a signal.
	// Probably because of resource error or memory violations.
	KilledBySignal = iota

	// Timeout means that the running program exceeded the target timeout.
	Timeout = iota

	// InternalError means that the Runner encountered an error.
	InternalError = iota
)

func (be BoxError) String() string {
	switch be {
	case NoError:
		return "NoError"
	case RunTimeError:
		return "RunTimeError"
	case KilledBySignal:
		return "KilledBySignal"
	case Timeout:
		return "Timeout"
	case InternalError:
		return "InternalError"
	default:
		return ""
	}
}

// Box represents an isolated environment
type Box struct {
	Config BoxConfig
	Path   string
	ID     uint

	runner Runner
}

// DefaultBoxConfig returns a new instance of the default box config
func DefaultBoxConfig() BoxConfig {
	bc := BoxConfig{}

	bc.Env = make([]EnvPair, 1)
	bc.Env[0].Var = "LIBC_FATAL_STDERR_"
	bc.Env[0].Value = "1"

	return bc
}

// NewBox returns a new Box instance
func NewBox() *Box {
	b := Box{}
	b.Config = DefaultBoxConfig()
	switch runtime.GOOS {
	case "linux":
		b.runner = &IsolateRunner{&b}
	case "freebsd":
		b.runner = &JailsRunner{&b}
	default:
		// XXX: replace with error
		panic("Unsupported OS")
	}

	return &b
}

// Init calls the runner's Init function.
func (b *Box) Init() error {
	return b.runner.Init()
}

// Run calls the runner's Run function
func (b *Box) Run(command string) (RunResult, error) {
	return b.runner.Run(command)
}

// Cleanup calls the runner's Cleanup function.
func (b *Box) Cleanup() error {
	return b.runner.Cleanup()
}
