package uitask_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/uitask"
)

// TestMarkInterruptedReportsCanceledNotFailed covers the distinction between a task that
// broke and one that was cut short by something outside it, such as the machine going to
// sleep mid-snapshot. Both end with a non-nil error; only the former should be presented
// to the user as a failure.
func TestMarkInterruptedReportsCanceledNotFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	boom := errors.New("boom")

	t.Run("interrupted task is canceled", func(t *testing.T) {
		t.Parallel()

		m := uitask.NewManager(false)

		var taskID string

		//nolint:errcheck
		m.Run(ctx, "Snapshot", "interrupted", func(ctx context.Context, ctrl uitask.Controller) error {
			taskID = ctrl.CurrentTaskID()
			ctrl.MarkInterrupted()

			return boom
		})

		got, ok := m.GetTask(taskID)
		if !ok {
			t.Fatalf("task %q not found", taskID)
		}

		if got.Status != uitask.StatusCanceled {
			t.Errorf("Status = %v, want %v", got.Status, uitask.StatusCanceled)
		}

		// the explanation must survive: it is what the user reads against the task.
		if got.ErrorMessage == "" {
			t.Error("ErrorMessage was dropped; the user is left with no explanation")
		}
	})

	t.Run("ordinary failure is still failed", func(t *testing.T) {
		t.Parallel()

		m := uitask.NewManager(false)

		var taskID string

		//nolint:errcheck
		m.Run(ctx, "Snapshot", "failed", func(ctx context.Context, ctrl uitask.Controller) error {
			taskID = ctrl.CurrentTaskID()

			return boom
		})

		got, ok := m.GetTask(taskID)
		if !ok {
			t.Fatalf("task %q not found", taskID)
		}

		if got.Status != uitask.StatusFailed {
			t.Errorf("Status = %v, want %v", got.Status, uitask.StatusFailed)
		}
	})

	t.Run("success is unaffected", func(t *testing.T) {
		t.Parallel()

		m := uitask.NewManager(false)

		var taskID string

		//nolint:errcheck
		m.Run(ctx, "Snapshot", "ok", func(ctx context.Context, ctrl uitask.Controller) error {
			taskID = ctrl.CurrentTaskID()

			return nil
		})

		got, ok := m.GetTask(taskID)
		if !ok {
			t.Fatalf("task %q not found", taskID)
		}

		if got.Status != uitask.StatusSuccess {
			t.Errorf("Status = %v, want %v", got.Status, uitask.StatusSuccess)
		}
	})
}
