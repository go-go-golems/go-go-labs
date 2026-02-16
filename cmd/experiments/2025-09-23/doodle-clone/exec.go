package main

import (
    "context"
    "encoding/json"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/pkg/errors"
)

type ExecutorOption func(*Executor)

func WithVerbose(v bool) ExecutorOption {
    return func(e *Executor) { e.verbose = v }
}

type Executor struct {
    store   *Store
    verbose bool
    ids     map[string]ActionResult
}

type ActionResult struct {
    InputID    string
    Operation  string
    PollID     string
    SlotIDs    []string
    EventID    string
    ConstrIDs  []string
    Candidates []Candidate
    Summary    string
}

type Candidate struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
    Score int       `json:"score"`
}

func NewExecutor(store *Store, opts ...ExecutorOption) *Executor {
    e := &Executor{
        store: store,
        ids:   map[string]ActionResult{},
    }
    for _, o := range opts {
        o(e)
    }
    return e
}

func (e *Executor) Run(ctx context.Context, doc *Document) ([]ActionResult, error) {
    var out []ActionResult
    for _, a := range doc.Actions {
        if a.When != "" && a.When != "now" {
            out = append(out, ActionResult{
                InputID:   a.ID,
                Operation: a.Action,
                Summary:   "skipped (when != now)",
            })
            continue
        }

        res, err := e.execOne(ctx, a)
        if err != nil {
            return out, errors.Wrapf(err, "action id=%q op=%s", a.ID, a.Action)
        }
        out = append(out, res)
        if a.ID != "" {
            e.ids[a.ID] = res
        }
    }
    return out, nil
}

func (e *Executor) execOne(ctx context.Context, a ActionPayload) (ActionResult, error) {
    loc, err := resolveTZ(a.UseTZ)
    if err != nil {
        return ActionResult{}, err
    }

    switch strings.ToLower(a.Action) {
    case "create_poll":
        return e.actCreatePoll(ctx, a, loc)
    case "add_slots":
        return e.actAddSlots(ctx, a, loc)
    case "vote_slot":
        return e.actVoteSlot(ctx, a, loc)
    case "finalize_poll":
        return e.actFinalizePoll(ctx, a)
    case "create_event":
        return e.actCreateEvent(ctx, a, loc)
    case "propose_times":
        return e.actProposeTimes(ctx, a, loc)
    case "reschedule_event":
        return e.actRescheduleEvent(ctx, a, loc)
    case "set_constraints":
        return e.actSetConstraints(ctx, a)
    case "remove_constraints":
        return e.actRemoveConstraints(ctx, a)
    case "sync_now":
        return ActionResult{
            InputID:   a.ID,
            Operation: a.Action,
            Summary:   "sync simulated (no-op in CLI)",
        }, nil
    default:
        return ActionResult{}, errors.Errorf("unsupported action: %s", a.Action)
    }
}

func (e *Executor) actCreatePoll(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    durationMin := 0
    if a.Duration != "" {
        d, err := time.ParseDuration(a.Duration)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse duration")
        }
        durationMin = int(d / time.Minute)
    }

    var deadlinePtr *time.Time
    if a.Deadline != "" {
        t, err := parseTime(a.Deadline, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse deadline")
        }
        deadlinePtr = &t
    }

    p := &Poll{
        Title:           a.Title,
        Strategy:        valOr(a.Strategy, "approval"),
        Quorum:          a.Quorum,
        Deadline:        deadlinePtr,
        Notes:           a.Notes,
        DurationMinutes: durationMin,
    }
    if err := e.store.CreatePoll(ctx, p); err != nil {
        return ActionResult{}, err
    }

    for _, pr := range a.Participants {
        pp := &PollParticipant{
            PollID: p.ID,
            Email:  pr.Email,
            Role:   pr.Role,
        }
        if err := e.store.AddPollParticipant(ctx, pp); err != nil {
            return ActionResult{}, err
        }
    }

    for _, w := range a.CandidateWindows {
        ws, err := parseTime(w.Start, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse window.start")
        }
        we, err := parseTime(w.End, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse window.end")
        }
        if !ws.Before(we) {
            return ActionResult{}, errors.Errorf("window has start >= end")
        }
        if err := e.store.AddPollWindow(ctx, &PollWindow{
            PollID: p.ID,
            Start:  ws.UTC(),
            End:    we.UTC(),
        }); err != nil {
            return ActionResult{}, err
        }
    }

    for _, c := range a.Constraints {
        payloadJSON, _ := json.Marshal(c.Payload)
        if err := e.store.InsertConstraint(ctx, &Constraint{
            Scope:       "poll",
            ScopeRef:    p.ID,
            Kind:        c.Kind,
            PayloadJSON: string(payloadJSON),
        }); err != nil {
            return ActionResult{}, errors.Wrap(err, "insert constraint")
        }
    }

    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        PollID:    p.ID,
        Summary:   fmt.Sprintf("poll created: %s", p.ID),
    }, nil
}

func (e *Executor) actAddSlots(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    pollID, err := e.resolvePollRef(a.PollRef)
    if err != nil {
        return ActionResult{}, err
    }
    var ids []string
    for _, s := range a.Slots {
        st, err := parseTime(s.Start, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse slot.start")
        }
        et, err := parseTime(s.End, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse slot.end")
        }
        sl := &Slot{
            PollID: pollID,
            Start:  st.UTC(),
            End:    et.UTC(),
        }
        if err := e.store.AddSlot(ctx, sl); err != nil {
            return ActionResult{}, err
        }
        ids = append(ids, sl.ID)
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        PollID:    pollID,
        SlotIDs:   ids,
        Summary:   fmt.Sprintf("slots added: %d", len(ids)),
    }, nil
}

func (e *Executor) actVoteSlot(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    pollID, err := e.resolvePollRef(a.PollRef)
    if err != nil {
        return ActionResult{}, err
    }
    var updated int
    for _, v := range a.Votes {
        slot, err := e.resolveSlotRef(ctx, pollID, v.SlotRef, loc)
        if err != nil {
            return ActionResult{}, err
        }
        if slot == nil {
            return ActionResult{}, errors.Errorf("slot not found for ref: %s", v.SlotRef)
        }
        email := v.Email
        if email == "" {
            email = "unknown@example.com"
        }
        err = e.store.UpsertVote(ctx, &Vote{
            PollID:  pollID,
            SlotID:  slot.ID,
            Email:   email,
            Vote:    strings.ToLower(v.Vote),
            Comment: v.Comment,
        })
        if err != nil {
            return ActionResult{}, err
        }
        updated++
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        PollID:    pollID,
        Summary:   fmt.Sprintf("votes recorded: %d", updated),
    }, nil
}

func (e *Executor) actFinalizePoll(ctx context.Context, a ActionPayload) (ActionResult, error) {
    pollID, err := e.resolvePollRef(a.PollRef)
    if err != nil {
        return ActionResult{}, err
    }
    slots, err := e.store.GetSlotsByPoll(ctx, pollID)
    if err != nil {
        return ActionResult{}, err
    }
    if len(slots) == 0 {
        return ActionResult{}, errors.Errorf("no slots to finalize for poll %s", pollID)
    }

    if len(a.PreferredOrder) > 0 {
        for _, id := range a.PreferredOrder {
            for _, s := range slots {
                if s.ID == id {
                    evt := &Event{
                        Title: fmt.Sprintf("Poll %s", pollID),
                        Start: s.Start,
                        End:   s.End,
                    }
                    if err := e.store.CreateEvent(ctx, evt); err != nil {
                        return ActionResult{}, err
                    }
                    if err := e.store.SetPollEvent(ctx, pollID, evt.ID); err != nil {
                        return ActionResult{}, err
                    }
                    return ActionResult{
                        InputID:   a.ID,
                        Operation: a.Action,
                        PollID:    pollID,
                        EventID:   evt.ID,
                        Summary:   fmt.Sprintf("finalized with preferred slot %s", s.ID),
                    }, nil
                }
            }
        }
    }

    type sc struct {
        Slot Slot
        Yes  int
    }
    var scored []sc
    for _, s := range slots {
        yes, err := e.store.CountYesVotes(ctx, s.ID)
        if err != nil {
            return ActionResult{}, err
        }
        scored = append(scored, sc{Slot: s, Yes: yes})
    }
    sort.Slice(scored, func(i, j int) bool {
        if scored[i].Yes == scored[j].Yes {
            return scored[i].Slot.Start.Before(scored[j].Slot.Start)
        }
        return scored[i].Yes > scored[j].Yes
    })
    chosen := scored[0].Slot

    evt := &Event{
        Title: fmt.Sprintf("Poll %s", pollID),
        Start: chosen.Start,
        End:   chosen.End,
    }
    if err := e.store.CreateEvent(ctx, evt); err != nil {
        return ActionResult{}, err
    }
    if err := e.store.SetPollEvent(ctx, pollID, evt.ID); err != nil {
        return ActionResult{}, err
    }

    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        PollID:    pollID,
        EventID:   evt.ID,
        Summary:   fmt.Sprintf("finalized with slot %s (yes=%d)", chosen.ID, scored[0].Yes),
    }, nil
}

func (e *Executor) actCreateEvent(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    st, err := parseTime(a.Start, loc)
    if err != nil {
        return ActionResult{}, errors.Wrap(err, "parse start")
    }
    en, err := parseTime(a.End, loc)
    if err != nil {
        return ActionResult{}, errors.Wrap(err, "parse end")
    }
    evt := &Event{
        Title:       a.Title,
        Start:       st.UTC(),
        End:         en.UTC(),
        Location:    a.Location,
        Notes:       a.Notes,
        CalendarRef: a.CalendarRef,
    }
    if err := e.store.CreateEvent(ctx, evt); err != nil {
        return ActionResult{}, err
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        EventID:   evt.ID,
        Summary:   fmt.Sprintf("event created: %s", evt.ID),
    }, nil
}

func (e *Executor) actProposeTimes(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    if len(a.CandidateWindows) == 0 {
        return ActionResult{}, errors.Errorf("propose_times requires candidate_windows")
    }
    dur, err := time.ParseDuration(valOr(a.Duration, "30m"))
    if err != nil {
        return ActionResult{}, errors.Wrap(err, "parse duration")
    }
    var out []Candidate
    for _, w := range a.CandidateWindows {
        ws, err := parseTime(w.Start, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse window.start")
        }
        we, err := parseTime(w.End, loc)
        if err != nil {
            return ActionResult{}, errors.Wrap(err, "parse window.end")
        }
        step := 30 * time.Minute
        for cur := ws; cur.Add(dur).Before(we) || cur.Add(dur).Equal(we); cur = cur.Add(step) {
            out = append(out, Candidate{
                Start: cur.UTC(),
                End:   cur.Add(dur).UTC(),
                Score: 0,
            })
            if a.MaxCandidates > 0 && len(out) >= a.MaxCandidates {
                break
            }
        }
        if a.MaxCandidates > 0 && len(out) >= a.MaxCandidates {
            break
        }
    }

    b, _ := json.MarshalIndent(out, "", "  ")
    fmt.Printf("propose_times candidates:\n%s\n", string(b))

    return ActionResult{
        InputID:    a.ID,
        Operation:  a.Action,
        Candidates: out,
        Summary:    fmt.Sprintf("candidates=%d", len(out)),
    }, nil
}

func (e *Executor) actRescheduleEvent(ctx context.Context, a ActionPayload, loc *time.Location) (ActionResult, error) {
    if a.EventRef == "" {
        return ActionResult{}, errors.Errorf("event_ref required")
    }
    if len(a.CandidateWindows) == 0 {
        return ActionResult{}, errors.Errorf("candidate_windows required")
    }
    ws, err := parseTime(a.CandidateWindows[0].Start, loc)
    if err != nil {
        return ActionResult{}, errors.Wrap(err, "parse candidate_windows[0].start")
    }
    we, err := parseTime(a.CandidateWindows[0].End, loc)
    if err != nil {
        return ActionResult{}, errors.Wrap(err, "parse candidate_windows[0].end")
    }
    if !ws.Before(we) {
        return ActionResult{}, errors.Errorf("window start >= end")
    }
    d := time.Hour
    if a.Duration != "" {
        if dd, err := time.ParseDuration(a.Duration); err == nil {
            d = dd
        }
    }
    newStart := ws.UTC()
    newEnd := ws.Add(d).UTC()
    if newEnd.After(we.UTC()) {
        return ActionResult{}, errors.Errorf("duration doesn't fit into window")
    }
    if err := e.store.UpdateEventTimes(ctx, a.EventRef, newStart, newEnd); err != nil {
        return ActionResult{}, err
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        EventID:   a.EventRef,
        Summary:   fmt.Sprintf("event %s rescheduled to %s - %s", a.EventRef, newStart.Format(time.RFC3339), newEnd.Format(time.RFC3339)),
    }, nil
}

func (e *Executor) actSetConstraints(ctx context.Context, a ActionPayload) (ActionResult, error) {
    var ids []string
    for _, c := range a.Constraints {
        payloadJSON, _ := json.Marshal(c.Payload)
        rc := &Constraint{
            Scope:       valOr(c.Scope, valOr(a.Scope, "user")),
            ScopeRef:    a.ScopeRef,
            Kind:        c.Kind,
            PayloadJSON: string(payloadJSON),
        }
        if err := e.store.InsertConstraint(ctx, rc); err != nil {
            return ActionResult{}, err
        }
        ids = append(ids, rc.ID)
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        ConstrIDs: ids,
        Summary:   fmt.Sprintf("constraints added: %d", len(ids)),
    }, nil
}

func (e *Executor) actRemoveConstraints(ctx context.Context, a ActionPayload) (ActionResult, error) {
    aff, err := e.store.DeleteConstraints(ctx, a.ConstraintIDs)
    if err != nil {
        return ActionResult{}, err
    }
    return ActionResult{
        InputID:   a.ID,
        Operation: a.Action,
        Summary:   fmt.Sprintf("constraints removed: %d", aff),
    }, nil
}

func (e *Executor) resolvePollRef(ref string) (string, error) {
    if ref == "" {
        return "", errors.Errorf("poll_ref required")
    }
    if res, ok := e.ids[ref]; ok && res.PollID != "" {
        return res.PollID, nil
    }
    return ref, nil
}

func (e *Executor) resolveSlotRef(ctx context.Context, pollID string, ref string, loc *time.Location) (*Slot, error) {
    if sl, err := e.store.GetSlotByID(ctx, ref); err == nil && sl != nil && sl.PollID == pollID {
        return sl, nil
    }
    t, err := parseTime(ref, loc)
    if err == nil {
        sl, err := e.store.GetSlotByStart(ctx, pollID, t.UTC())
        if err != nil {
            return nil, err
        }
        return sl, nil
    }
    return nil, errors.Errorf("cannot resolve slot_ref: %s", ref)
}

func resolveTZ(tz string) (*time.Location, error) {
    if tz == "" {
        return time.Local, nil
    }
    loc, err := time.LoadLocation(tz)
    if err != nil {
        return nil, errors.Wrap(err, "load tz")
    }
    return loc, nil
}

func parseTime(s string, loc *time.Location) (time.Time, error) {
    if t, err := time.Parse(time.RFC3339, s); err == nil {
        return t, nil
    }
    layouts := []string{
        "2006-01-02T15:04:05",
        "2006-01-02T15:04",
        "2006-01-02 15:04:05",
        "2006-01-02 15:04",
    }
    for _, l := range layouts {
        if t, err := time.ParseInLocation(l, s, loc); err == nil {
            return t, nil
        }
    }
    return time.Time{}, errors.Errorf("unsupported time format: %q", s)
}

func valOr[T ~string](v T, def T) T {
    if string(v) == "" {
        return def
    }
    return v
}


