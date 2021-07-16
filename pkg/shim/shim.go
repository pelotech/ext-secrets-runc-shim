package shim

import (
	"context"
	"os"

	v2 "github.com/containerd/containerd/runtime/v2/runc/v2"
	"github.com/containerd/containerd/runtime/v2/shim"
	taskAPI "github.com/containerd/containerd/runtime/v2/task"
	ptypes "github.com/gogo/protobuf/types"
)

var (
	// check to make sure the *service implements the GRPC API
	_ = (taskAPI.TaskService)(&service{})
)

// New returns a new shim service
func New(ctx context.Context, id string, publisher shim.Publisher, shutdown func()) (shim.Shim, error) {
	var err error

	s := &service{}
	s.runc, err = v2.New(ctx, id, publisher, shutdown)
	if err != nil {
		return nil, err
	}
	s.curdir, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	return s, nil
}

type service struct {
	runc shim.Shim

	curdir string
}

// StartShim is a binary call that executes a new shim returning the address
func (s *service) StartShim(ctx context.Context, opts shim.StartOpts) (string, error) {
	// TODO: underlying shim should be reimplemented so secret replacement
	// can happen later. The runc shim caches the config after this call
	// and most of the code is private to the package.
	spec, err := s.getSpec(ctx)
	if err != nil {
		return "", err
	}
	if err := parseSpec(ctx, spec); err != nil {
		return "", err
	}
	if err := s.writeSpec(ctx, spec); err != nil {
		return "", err
	}
	return s.runc.StartShim(ctx, opts)
}

// Cleanup is a binary call that cleans up any resources used by the shim when the service crashes
func (s *service) Cleanup(ctx context.Context) (*taskAPI.DeleteResponse, error) {
	return s.runc.Cleanup(ctx)
}

// Create a new container
func (s *service) Create(ctx context.Context, r *taskAPI.CreateTaskRequest) (_ *taskAPI.CreateTaskResponse, err error) {
	return s.runc.Create(ctx, r)
}

// Start the primary user process inside the container
func (s *service) Start(ctx context.Context, r *taskAPI.StartRequest) (*taskAPI.StartResponse, error) {
	return s.runc.Start(ctx, r)
}

// Delete a process or container
func (s *service) Delete(ctx context.Context, r *taskAPI.DeleteRequest) (*taskAPI.DeleteResponse, error) {
	return s.runc.Delete(ctx, r)
}

// Exec an additional process inside the container
func (s *service) Exec(ctx context.Context, r *taskAPI.ExecProcessRequest) (*ptypes.Empty, error) {
	return s.runc.Exec(ctx, r)
}

// ResizePty of a process
func (s *service) ResizePty(ctx context.Context, r *taskAPI.ResizePtyRequest) (*ptypes.Empty, error) {
	return s.runc.ResizePty(ctx, r)
}

// State returns runtime state of a process
func (s *service) State(ctx context.Context, r *taskAPI.StateRequest) (*taskAPI.StateResponse, error) {
	return s.runc.State(ctx, r)
}

// Pause the container
func (s *service) Pause(ctx context.Context, r *taskAPI.PauseRequest) (*ptypes.Empty, error) {
	return s.runc.Pause(ctx, r)
}

// Resume the container
func (s *service) Resume(ctx context.Context, r *taskAPI.ResumeRequest) (*ptypes.Empty, error) {
	return s.runc.Resume(ctx, r)
}

// Kill a process
func (s *service) Kill(ctx context.Context, r *taskAPI.KillRequest) (*ptypes.Empty, error) {
	return s.runc.Kill(ctx, r)
}

// Pids returns all pids inside the container
func (s *service) Pids(ctx context.Context, r *taskAPI.PidsRequest) (*taskAPI.PidsResponse, error) {
	return s.runc.Pids(ctx, r)
}

// CloseIO of a process
func (s *service) CloseIO(ctx context.Context, r *taskAPI.CloseIORequest) (*ptypes.Empty, error) {
	return s.runc.CloseIO(ctx, r)
}

// Checkpoint the container
func (s *service) Checkpoint(ctx context.Context, r *taskAPI.CheckpointTaskRequest) (*ptypes.Empty, error) {
	return s.runc.Checkpoint(ctx, r)
}

// Connect returns shim information of the underlying service
func (s *service) Connect(ctx context.Context, r *taskAPI.ConnectRequest) (*taskAPI.ConnectResponse, error) {
	return s.runc.Connect(ctx, r)
}

// Shutdown is called after the underlying resources of the shim are cleaned up and the service can be stopped
func (s *service) Shutdown(ctx context.Context, r *taskAPI.ShutdownRequest) (*ptypes.Empty, error) {
	return s.runc.Shutdown(ctx, r)
}

// Stats returns container level system stats for a container and its processes
func (s *service) Stats(ctx context.Context, r *taskAPI.StatsRequest) (*taskAPI.StatsResponse, error) {
	return s.runc.Stats(ctx, r)
}

// Update the live container
func (s *service) Update(ctx context.Context, r *taskAPI.UpdateTaskRequest) (*ptypes.Empty, error) {
	return s.runc.Update(ctx, r)
}

// Wait for a process to exit
func (s *service) Wait(ctx context.Context, r *taskAPI.WaitRequest) (*taskAPI.WaitResponse, error) {
	return s.runc.Wait(ctx, r)
}
