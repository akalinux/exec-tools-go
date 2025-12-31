package cmd

// https://stackoverflow.com/questions/23031752/start-a-process-in-go-and-detach-from-it
// https://cs.opensource.google/go/go/+/refs/tags/go1.25.5:src/os/exec/exec.go
import (
	"errors"
	"os"
	"syscall"
)

type Cmd struct {
	Name        string
	Args        []string
	Process     *os.Process
	SysProcAttr *syscall.SysProcAttr
	Stdin       *RwSet
	Stdout      *RwSet
	Stderr      *RwSet
	Dir         string
	Env         []string
}

func NewCmd(name string, args ...string) *Cmd {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return &Cmd{
		Name:        name,
		Args:        args,
		SysProcAttr: DefaultSysProcAttr(),
		Dir:         dir,

		Env:    os.Environ(),
		Stdin:  &RwSet{},
		Stdout: &RwSet{},
		Stderr: &RwSet{},
	}
}

func (s *Cmd) newPs(set *RwSet) (*os.File, error) {
	if set.Read != nil || set.Write != nil {
		return nil, errors.New("Reader/Writer all ready defined!")
	}
	r, w, e := os.Pipe()
	if e != nil {
		return nil, e
	}
	set.Read = r
	set.Write = w
	return w, nil
}

func (s *Cmd) NewStdin() (*os.File, error) {
	return s.newPs(s.Stdin)
}

func (s *Cmd) NewStdout() (*os.File, error) {
	return s.newPs(s.Stdout)
}

func (s *Cmd) NewStderr() (*os.File, error) {
	return s.newPs(s.Stderr)
}

// Generates the our default SysProcAttr.
func DefaultSysProcAttr() *syscall.SysProcAttr {
	glist, err := os.Getgroups()

	groups := make([]uint32, len(glist))
	if err != nil {
		for i, group := range glist {
			groups[i] = uint32(group)
		}
	}

	cred := &syscall.Credential{
		Uid:    uint32(os.Getegid()),
		Gid:    uint32(os.Getegid()),
		Groups: groups,
	}
	return &syscall.SysProcAttr{
		Credential: cred,
		Setsid:     true, // detach by default
		Pdeathsig:  syscall.SIGTERM,
		Noctty:     true,
	}
}
