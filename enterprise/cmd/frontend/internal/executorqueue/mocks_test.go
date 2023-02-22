// Code generated by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package executorqueue

import (
	"sync"

	api "github.com/sourcegraph/sourcegraph/internal/api"
)

// MockGitserverClient is a mock implementation of the GitserverClient
// interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue)
// used for unit testing.
type MockGitserverClient struct {
	// AddrForRepoFunc is an instance of a mock function object controlling
	// the behavior of the method AddrForRepo.
	AddrForRepoFunc *GitserverClientAddrForRepoFunc
}

// NewMockGitserverClient creates a new mock of the GitserverClient
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defaultHook: func(api.RepoName) (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockGitserverClient creates a new mock of the GitserverClient
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defaultHook: func(api.RepoName) string {
				panic("unexpected invocation of MockGitserverClient.AddrForRepo")
			},
		},
	}
}

// NewMockGitserverClientFrom creates a new mock of the MockGitserverClient
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockGitserverClientFrom(i GitserverClient) *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defaultHook: i.AddrForRepo,
		},
	}
}

// GitserverClientAddrForRepoFunc describes the behavior when the
// AddrForRepo method of the parent MockGitserverClient instance is invoked.
type GitserverClientAddrForRepoFunc struct {
	defaultHook func(api.RepoName) string
	hooks       []func(api.RepoName) string
	history     []GitserverClientAddrForRepoFuncCall
	mutex       sync.Mutex
}

// AddrForRepo delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockGitserverClient) AddrForRepo(v0 api.RepoName) string {
	r0 := m.AddrForRepoFunc.nextHook()(v0)
	m.AddrForRepoFunc.appendCall(GitserverClientAddrForRepoFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the AddrForRepo method
// of the parent MockGitserverClient instance is invoked and the hook queue
// is empty.
func (f *GitserverClientAddrForRepoFunc) SetDefaultHook(hook func(api.RepoName) string) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// AddrForRepo method of the parent MockGitserverClient instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *GitserverClientAddrForRepoFunc) PushHook(hook func(api.RepoName) string) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *GitserverClientAddrForRepoFunc) SetDefaultReturn(r0 string) {
	f.SetDefaultHook(func(api.RepoName) string {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *GitserverClientAddrForRepoFunc) PushReturn(r0 string) {
	f.PushHook(func(api.RepoName) string {
		return r0
	})
}

func (f *GitserverClientAddrForRepoFunc) nextHook() func(api.RepoName) string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientAddrForRepoFunc) appendCall(r0 GitserverClientAddrForRepoFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientAddrForRepoFuncCall objects
// describing the invocations of this function.
func (f *GitserverClientAddrForRepoFunc) History() []GitserverClientAddrForRepoFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientAddrForRepoFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientAddrForRepoFuncCall is an object that describes an
// invocation of method AddrForRepo on an instance of MockGitserverClient.
type GitserverClientAddrForRepoFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 api.RepoName
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientAddrForRepoFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientAddrForRepoFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
