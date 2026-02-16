package main

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
    "github.com/pkg/errors"
)

type Store struct {
    db *sqlx.DB
}

func OpenStore(ctx context.Context, path string) (*Store, error) {
    db, err := sqlx.ConnectContext(ctx, "sqlite3", path+"?_foreign_keys=1&_journal_mode=WAL")
    if err != nil {
        return nil, errors.Wrap(err, "connect sqlite")
    }
    return &Store{db: db}, nil
}

func (s *Store) Close() error {
    if s.db == nil {
        return nil
    }
    return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
    schema := `
CREATE TABLE IF NOT EXISTS polls (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  strategy TEXT NOT NULL DEFAULT 'approval',
  quorum INTEGER NOT NULL DEFAULT 0,
  deadline DATETIME,
  notes TEXT,
  duration_minutes INTEGER,
  status TEXT NOT NULL DEFAULT 'draft', -- draft|finalized
  event_id TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS poll_participants (
  id TEXT PRIMARY KEY,
  poll_id TEXT NOT NULL,
  email TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'required',
  UNIQUE(poll_id, email),
  FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS poll_windows (
  id TEXT PRIMARY KEY,
  poll_id TEXT NOT NULL,
  start_ts DATETIME NOT NULL,
  end_ts DATETIME NOT NULL,
  FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS slots (
  id TEXT PRIMARY KEY,
  poll_id TEXT NOT NULL,
  start_ts DATETIME NOT NULL,
  end_ts DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(poll_id, start_ts),
  FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS votes (
  id TEXT PRIMARY KEY,
  poll_id TEXT NOT NULL,
  slot_id TEXT NOT NULL,
  email TEXT NOT NULL,
  vote TEXT NOT NULL, -- yes|no|maybe
  comment TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(slot_id, email),
  FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE,
  FOREIGN KEY (slot_id) REFERENCES slots(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  start_ts DATETIME NOT NULL,
  end_ts DATETIME NOT NULL,
  location TEXT,
  notes TEXT,
  calendar_ref TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS constraints (
  id TEXT PRIMARY KEY,
  scope TEXT NOT NULL, -- user|event|poll
  scope_ref TEXT,      -- nullable for user
  kind TEXT NOT NULL,
  payload TEXT NOT NULL, -- JSON
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_slots_poll ON slots(poll_id);
CREATE INDEX IF NOT EXISTS idx_votes_slot ON votes(slot_id);
CREATE INDEX IF NOT EXISTS idx_votes_poll ON votes(poll_id);
`
    _, err := s.db.ExecContext(ctx, schema)
    return errors.Wrap(err, "create schema")
}

func (s *Store) CreatePoll(ctx context.Context, p *Poll) error {
    if p.ID == "" {
        p.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO polls (id, title, strategy, quorum, deadline, notes, duration_minutes, status)
VALUES (?, ?, ?, ?, ?, ?, ?, 'draft')`,
        p.ID, p.Title, p.Strategy, p.Quorum, nullTime(p.Deadline), p.Notes, p.DurationMinutes)
    return errors.Wrap(err, "insert poll")
}

func (s *Store) SetPollEvent(ctx context.Context, pollID, eventID string) error {
    _, err := s.db.ExecContext(ctx, `UPDATE polls SET status='finalized', event_id=? WHERE id=?`, eventID, pollID)
    return errors.Wrap(err, "update poll finalized")
}

func (s *Store) AddPollParticipant(ctx context.Context, pp *PollParticipant) error {
    if pp.ID == "" {
        pp.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO poll_participants (id, poll_id, email, role)
VALUES (?, ?, ?, ?) ON CONFLICT(poll_id, email) DO UPDATE SET role=excluded.role`,
        pp.ID, pp.PollID, pp.Email, pp.RoleOrDefault())
    return errors.Wrap(err, "insert poll participant")
}

func (s *Store) AddPollWindow(ctx context.Context, w *PollWindow) error {
    if w.ID == "" {
        w.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO poll_windows (id, poll_id, start_ts, end_ts)
VALUES (?, ?, ?, ?)`,
        w.ID, w.PollID, w.Start, w.End)
    return errors.Wrap(err, "insert poll window")
}

func (s *Store) AddSlot(ctx context.Context, sl *Slot) error {
    if sl.ID == "" {
        sl.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO slots (id, poll_id, start_ts, end_ts)
VALUES (?, ?, ?, ?)`,
        sl.ID, sl.PollID, sl.Start, sl.End)
    return errors.Wrap(err, "insert slot")
}

func (s *Store) GetSlotsByPoll(ctx context.Context, pollID string) ([]Slot, error) {
    var out []Slot
    err := s.db.SelectContext(ctx, &out, `SELECT id, poll_id, start_ts, end_ts, created_at FROM slots WHERE poll_id=? ORDER BY start_ts ASC`, pollID)
    return out, errors.Wrap(err, "select slots by poll")
}

func (s *Store) GetSlotByID(ctx context.Context, slotID string) (*Slot, error) {
    var out Slot
    err := s.db.GetContext(ctx, &out, `SELECT id, poll_id, start_ts, end_ts, created_at FROM slots WHERE id=?`, slotID)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &out, errors.Wrap(err, "select slot by id")
}

func (s *Store) GetSlotByStart(ctx context.Context, pollID string, start time.Time) (*Slot, error) {
    var out Slot
    err := s.db.GetContext(ctx, &out, `SELECT id, poll_id, start_ts, end_ts, created_at FROM slots WHERE poll_id=? AND start_ts=?`, pollID, start)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &out, errors.Wrap(err, "select slot by start")
}

func (s *Store) UpsertVote(ctx context.Context, v *Vote) error {
    if v.ID == "" {
        v.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO votes (id, poll_id, slot_id, email, vote, comment)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(slot_id, email) DO UPDATE SET vote=excluded.vote, comment=excluded.comment`,
        v.ID, v.PollID, v.SlotID, v.Email, v.Vote, v.Comment)
    return errors.Wrap(err, "upsert vote")
}

func (s *Store) CountYesVotes(ctx context.Context, slotID string) (int, error) {
    var c int
    err := s.db.GetContext(ctx, &c, `SELECT COUNT(*) FROM votes WHERE slot_id=? AND vote='yes'`, slotID)
    return c, errors.Wrap(err, "count yes")
}

func (s *Store) CreateEvent(ctx context.Context, e *Event) error {
    if e.ID == "" {
        e.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO events (id, title, start_ts, end_ts, location, notes, calendar_ref)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
        e.ID, e.Title, e.Start, e.End, e.Location, e.Notes, e.CalendarRef)
    return errors.Wrap(err, "insert event")
}

func (s *Store) UpdateEventTimes(ctx context.Context, id string, start, end time.Time) error {
    _, err := s.db.ExecContext(ctx, `UPDATE events SET start_ts=?, end_ts=? WHERE id=?`, start, end, id)
    return errors.Wrap(err, "update event times")
}

func (s *Store) InsertConstraint(ctx context.Context, c *Constraint) error {
    if c.ID == "" {
        c.ID = uuid.NewString()
    }
    _, err := s.db.ExecContext(ctx, `
INSERT INTO constraints (id, scope, scope_ref, kind, payload)
VALUES (?, ?, ?, ?, ?)`,
        c.ID, c.Scope, c.ScopeRef, c.Kind, c.PayloadJSON)
    return errors.Wrap(err, "insert constraint")
}

func (s *Store) DeleteConstraints(ctx context.Context, ids []string) (int64, error) {
    if len(ids) == 0 {
        return 0, nil
    }
    q, args, err := sqlx.In(`DELETE FROM constraints WHERE id IN (?)`, ids)
    if err != nil {
        return 0, errors.Wrap(err, "build IN query")
    }
    q = s.db.Rebind(q)
    res, err := s.db.ExecContext(ctx, q, args...)
    if err != nil {
        return 0, errors.Wrap(err, "exec delete")
    }
    aff, _ := res.RowsAffected()
    return aff, nil
}

type Poll struct {
    ID              string
    Title           string
    Strategy        string
    Quorum          int
    Deadline        *time.Time
    Notes           string
    DurationMinutes int
}

type PollParticipant struct {
    ID     string
    PollID string
    Email  string
    Role   string
}

func (pp *PollParticipant) RoleOrDefault() string {
    if pp.Role == "" {
        return "required"
    }
    return pp.Role
}

type PollWindow struct {
    ID     string
    PollID string
    Start  time.Time
    End    time.Time
}

type Slot struct {
    ID     string    `db:"id"`
    PollID string    `db:"poll_id"`
    Start  time.Time `db:"start_ts"`
    End    time.Time `db:"end_ts"`
    CreatedAt time.Time `db:"created_at"`
}

type Vote struct {
    ID      string
    PollID  string
    SlotID  string
    Email   string
    Vote    string
    Comment string
}

type Event struct {
    ID          string
    Title       string
    Start       time.Time
    End         time.Time
    Location    string
    Notes       string
    CalendarRef string
}

type Constraint struct {
    ID          string
    Scope       string
    ScopeRef    string
    Kind        string
    PayloadJSON string
}

func nullTime(t *time.Time) interface{} {
    if t == nil {
        return nil
    }
    return *t
}


