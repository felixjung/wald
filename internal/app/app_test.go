package app

import (
	"context"
)

type runnerCall struct {
	Dir  string
	Name string
	Args []string
}

type fakeRunner struct {
	calls []runnerCall
	errs  map[int]error
}

func (f *fakeRunner) Run(_ context.Context, dir, name string, args ...string) error {
	current := len(f.calls)
	f.calls = append(f.calls, runnerCall{
		Dir:  dir,
		Name: name,
		Args: append([]string(nil), args...),
	})
	if err, ok := f.errs[current]; ok {
		return err
	}
	return nil
}
