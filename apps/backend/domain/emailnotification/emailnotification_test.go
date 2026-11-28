package emailnotification_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/emailnotification"
)

var baseTime = time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC)

func newPending(t *testing.T) emailnotification.EmailNotification {
	t.Helper()
	recipient, err := emailnotification.NewRecipientEmail("sakura@example.com")
	if err != nil {
		t.Fatal(err)
	}
	return emailnotification.New(
		uuid.New(), uuid.New(), recipient, kindReset,
		emailnotification.NewTemplateParams(map[string]string{keyURL: testURL}),
		baseTime,
	)
}

// claimTimes claims n repeatedly, as if each earlier attempt failed
// transiently, and returns the result of the last claim.
func claimTimes(t *testing.T, n emailnotification.EmailNotification, times int) emailnotification.EmailNotification {
	t.Helper()
	now := baseTime
	for i := range times {
		var err error
		n, err = n.Claim(now)
		if err != nil {
			t.Fatalf("Claim #%d error = %v", i+1, err)
		}
		if i < times-1 {
			n, err = n.RecordFailure(emailnotification.FailureTimeout, now)
			if err != nil {
				t.Fatalf("RecordFailure #%d error = %v", i+1, err)
			}
		}
		now = now.Add(time.Minute)
	}
	return n
}

func assertTerminal(t *testing.T, n emailnotification.EmailNotification, status emailnotification.EmailStatus) {
	t.Helper()
	if n.Status() != status {
		t.Fatalf("Status() = %s, want %s", n.Status(), status)
	}
	if _, ok := n.NextAttemptAt(); ok {
		t.Error("NextAttemptAt() is set in a terminal state")
	}
	if !n.TemplateParams().IsEmpty() {
		t.Error("TemplateParams() is not empty in a terminal state")
	}
}

func TestNew_initialState(t *testing.T) {
	id, userID := uuid.New(), uuid.New()
	recipient, _ := emailnotification.NewRecipientEmail("sakura@example.com")
	params := emailnotification.NewTemplateParams(map[string]string{keyURL: testURL})

	n := emailnotification.New(id, userID, recipient, kindReset, params, baseTime)

	if n.ID() != id || n.UserID() != userID {
		t.Errorf("ID/UserID = %v/%v, want %v/%v", n.ID(), n.UserID(), id, userID)
	}
	if n.Recipient() != recipient {
		t.Errorf("Recipient() = %v", n.Recipient())
	}
	if n.Kind() != kindReset {
		t.Errorf("Kind() = %v", n.Kind())
	}
	if n.Status() != emailnotification.StatusPending {
		t.Errorf("Status() = %v, want pending", n.Status())
	}
	if n.AttemptCount() != 0 {
		t.Errorf("AttemptCount() = %d, want 0", n.AttemptCount())
	}
	if !n.RequestedAt().Equal(baseTime) {
		t.Errorf("RequestedAt() = %v, want %v", n.RequestedAt(), baseTime)
	}
	if next, ok := n.NextAttemptAt(); !ok || !next.Equal(baseTime) {
		t.Errorf("NextAttemptAt() = %v, %v, want %v, true", next, ok, baseTime)
	}
	if _, ok := n.SentAt(); ok {
		t.Error("SentAt() is set")
	}
	if _, ok := n.LastFailureCode(); ok {
		t.Error("LastFailureCode() is set")
	}
	if n.TemplateParams().Values()[keyURL] != testURL {
		t.Errorf("TemplateParams() = %v", n.TemplateParams().Values())
	}
}

func TestClaim_incrementsAttemptAndLeases(t *testing.T) {
	n := newPending(t)
	now := baseTime.Add(5 * time.Second)

	claimed, err := n.Claim(now)
	if err != nil {
		t.Fatalf("Claim error = %v", err)
	}
	if claimed.Status() != emailnotification.StatusPending {
		t.Errorf("Status() = %v, want pending", claimed.Status())
	}
	if claimed.AttemptCount() != 1 {
		t.Errorf("AttemptCount() = %d, want 1", claimed.AttemptCount())
	}
	if next, _ := claimed.NextAttemptAt(); !next.Equal(now.Add(60 * time.Second)) {
		t.Errorf("NextAttemptAt() = %v, want now+60s", next)
	}
	if n.AttemptCount() != 0 {
		t.Error("Claim mutated the receiver")
	}
}

func TestClaim_failsAsWorkerInterruptedAfterMaxAttempts(t *testing.T) {
	n := claimTimes(t, newPending(t), 3)
	// The third attempt's worker stopped before recording a result.
	now := baseTime.Add(time.Hour)

	got, err := n.Claim(now)
	if err != nil {
		t.Fatalf("Claim error = %v", err)
	}
	assertTerminal(t, got, emailnotification.StatusFailed)
	if got.AttemptCount() != 3 {
		t.Errorf("AttemptCount() = %d, want 3 (not incremented)", got.AttemptCount())
	}
	if code, _ := got.LastFailureCode(); code != emailnotification.FailureWorkerInterrupted {
		t.Errorf("LastFailureCode() = %v, want worker_interrupted", code)
	}
}

func TestMarkSent(t *testing.T) {
	n := claimTimes(t, newPending(t), 1)
	now := baseTime.Add(3 * time.Second)

	sent, err := n.MarkSent(now)
	if err != nil {
		t.Fatalf("MarkSent error = %v", err)
	}
	assertTerminal(t, sent, emailnotification.StatusSent)
	if at, ok := sent.SentAt(); !ok || !at.Equal(now) {
		t.Errorf("SentAt() = %v, %v, want %v", at, ok, now)
	}
	if sent.AttemptCount() != 1 {
		t.Errorf("AttemptCount() = %d, want 1", sent.AttemptCount())
	}
}

func TestRecordFailure_transientSchedulesRetryWithBackoff(t *testing.T) {
	tests := []struct {
		attempts  int
		wantDelay time.Duration
	}{
		{1, 10 * time.Second},
		{2, 20 * time.Second},
	}
	for _, tt := range tests {
		n := claimTimes(t, newPending(t), tt.attempts)
		now := baseTime.Add(time.Hour)

		got, err := n.RecordFailure(emailnotification.FailureRateLimited, now)
		if err != nil {
			t.Fatalf("attempt %d: RecordFailure error = %v", tt.attempts, err)
		}
		if got.Status() != emailnotification.StatusPending {
			t.Fatalf("attempt %d: Status() = %v, want pending", tt.attempts, got.Status())
		}
		if next, _ := got.NextAttemptAt(); !next.Equal(now.Add(tt.wantDelay)) {
			t.Errorf("attempt %d: NextAttemptAt() = %v, want now+%v", tt.attempts, next, tt.wantDelay)
		}
		if code, _ := got.LastFailureCode(); code != emailnotification.FailureRateLimited {
			t.Errorf("attempt %d: LastFailureCode() = %v", tt.attempts, code)
		}
		if got.TemplateParams().IsEmpty() {
			t.Errorf("attempt %d: TemplateParams() cleared while pending", tt.attempts)
		}
		if got.AttemptCount() != tt.attempts {
			t.Errorf("attempt %d: AttemptCount() = %d", tt.attempts, got.AttemptCount())
		}
	}
}

func TestRecordFailure_transientOnLastAttemptFails(t *testing.T) {
	n := claimTimes(t, newPending(t), 3)

	got, err := n.RecordFailure(emailnotification.FailureConnectionFailed, baseTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("RecordFailure error = %v", err)
	}
	assertTerminal(t, got, emailnotification.StatusFailed)
	if code, _ := got.LastFailureCode(); code != emailnotification.FailureConnectionFailed {
		t.Errorf("LastFailureCode() = %v", code)
	}
}

func TestRecordFailure_permanentFailsImmediately(t *testing.T) {
	for _, code := range []emailnotification.FailureCode{
		emailnotification.FailureRecipientRejected,
		emailnotification.FailureTemplateRenderFailed,
	} {
		n := claimTimes(t, newPending(t), 1)
		got, err := n.RecordFailure(code, baseTime.Add(time.Second))
		if err != nil {
			t.Fatalf("%s: RecordFailure error = %v", code, err)
		}
		assertTerminal(t, got, emailnotification.StatusFailed)
		if c, _ := got.LastFailureCode(); c != code {
			t.Errorf("%s: LastFailureCode() = %v", code, c)
		}
	}
}

func TestTransitions_rejectTerminalStates(t *testing.T) {
	sent, err := claimTimes(t, newPending(t), 1).MarkSent(baseTime)
	if err != nil {
		t.Fatal(err)
	}
	failed, err := claimTimes(t, newPending(t), 1).RecordFailure(emailnotification.FailureAuthFailed, baseTime)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []emailnotification.EmailNotification{sent, failed} {
		if _, err := n.Claim(baseTime); !errors.Is(err, emailnotification.ErrNotPending) {
			t.Errorf("%s: Claim error = %v, want ErrNotPending", n.Status(), err)
		}
		if _, err := n.MarkSent(baseTime); !errors.Is(err, emailnotification.ErrNotPending) {
			t.Errorf("%s: MarkSent error = %v, want ErrNotPending", n.Status(), err)
		}
		_, err := n.RecordFailure(emailnotification.FailureTimeout, baseTime)
		if !errors.Is(err, emailnotification.ErrNotPending) {
			t.Errorf("%s: RecordFailure error = %v, want ErrNotPending", n.Status(), err)
		}
	}
}

func TestResultTransitions_requireClaim(t *testing.T) {
	n := newPending(t)
	if _, err := n.MarkSent(baseTime); !errors.Is(err, emailnotification.ErrNotClaimed) {
		t.Errorf("MarkSent error = %v, want ErrNotClaimed", err)
	}
	_, err := n.RecordFailure(emailnotification.FailureTimeout, baseTime)
	if !errors.Is(err, emailnotification.ErrNotClaimed) {
		t.Errorf("RecordFailure error = %v, want ErrNotClaimed", err)
	}
}

func validRestoreParams(t *testing.T) emailnotification.RestoreParams {
	t.Helper()
	recipient, _ := emailnotification.NewRecipientEmail("sakura@example.com")
	next := baseTime
	return emailnotification.RestoreParams{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Recipient:      recipient,
		Kind:           kindReset,
		Status:         emailnotification.StatusPending,
		AttemptCount:   1,
		RequestedAt:    baseTime,
		NextAttemptAt:  &next,
		TemplateParams: emailnotification.NewTemplateParams(map[string]string{keyURL: "x"}),
	}
}

func TestRestore_returnsStoredFields(t *testing.T) {
	p := validRestoreParams(t)
	code := emailnotification.FailureTimeout
	p.LastFailureCode = &code

	n, err := emailnotification.Restore(p)
	if err != nil {
		t.Fatalf("Restore error = %v", err)
	}
	if n.ID() != p.ID || n.AttemptCount() != 1 || n.Status() != emailnotification.StatusPending {
		t.Errorf("Restore returned %v/%d/%s", n.ID(), n.AttemptCount(), n.Status())
	}
	if c, ok := n.LastFailureCode(); !ok || c != code {
		t.Errorf("LastFailureCode() = %v, %v", c, ok)
	}
	sentAt := baseTime
	sent := validRestoreParams(t)
	sent.Status = emailnotification.StatusSent
	sent.NextAttemptAt = nil
	sent.SentAt = &sentAt
	sent.TemplateParams = emailnotification.NewTemplateParams(nil)
	if _, err := emailnotification.Restore(sent); err != nil {
		t.Errorf("Restore(sent) error = %v", err)
	}
}

func TestRestore_rejectsBrokenInvariants(t *testing.T) {
	ts := baseTime
	code := emailnotification.FailureAuthFailed
	empty := emailnotification.NewTemplateParams(nil)
	tests := map[string]func(p *emailnotification.RestoreParams){
		"pending without next attempt": func(p *emailnotification.RestoreParams) { p.NextAttemptAt = nil },
		"sent with next attempt": func(p *emailnotification.RestoreParams) {
			p.Status, p.SentAt, p.TemplateParams = emailnotification.StatusSent, &ts, empty
		},
		"sent without sent_at": func(p *emailnotification.RestoreParams) {
			p.Status, p.NextAttemptAt, p.TemplateParams = emailnotification.StatusSent, nil, empty
		},
		"failed without failure code": func(p *emailnotification.RestoreParams) {
			p.Status, p.NextAttemptAt, p.TemplateParams = emailnotification.StatusFailed, nil, empty
		},
		"terminal with params": func(p *emailnotification.RestoreParams) {
			p.Status, p.NextAttemptAt, p.LastFailureCode = emailnotification.StatusFailed, nil, &code
		},
		"negative attempt count": func(p *emailnotification.RestoreParams) { p.AttemptCount = -1 },
		"unknown status":         func(p *emailnotification.RestoreParams) { p.Status = "queued" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			p := validRestoreParams(t)
			mutate(&p)
			if _, err := emailnotification.Restore(p); err == nil {
				t.Fatal("Restore error = nil, want error")
			}
		})
	}
}
