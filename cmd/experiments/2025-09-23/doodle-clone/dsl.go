package main

import (
    "io"

    "gopkg.in/yaml.v3"
)

const DSLVersion = "doodle-v1"

type Document struct {
    Version string          `yaml:"version"`
    Actions []ActionPayload `yaml:"actions"`
}

type ActionPayload struct {
    // common envelope
    ID    string `yaml:"id"`
    When  string `yaml:"when"`   // now|on_approve (we execute only "now")
    UseTZ string `yaml:"use_tz"` // e.g. "America/New_York"

    Action string `yaml:"action"`

    // create_poll
    Title            string            `yaml:"title"`
    Participants     []ParticipantYAML `yaml:"participants"`
    Duration         string            `yaml:"duration"`
    CandidateWindows []WindowYAML      `yaml:"candidate_windows"`
    Strategy         string            `yaml:"strategy"`
    Quorum           int               `yaml:"quorum"`
    Deadline         string            `yaml:"deadline"`
    Notes            string            `yaml:"notes"`
    Constraints      []ConstraintYAML  `yaml:"constraints"`

    // add_slots
    PollRef string     `yaml:"poll_ref"`
    Slots   []SlotYAML `yaml:"slots"`

    // vote_slot
    Votes []VoteYAML `yaml:"votes"`

    // finalize_poll
    PreferredOrder []string `yaml:"preferred_order"`

    // create_event
    Start       string `yaml:"start"`
    End         string `yaml:"end"`
    Location    string `yaml:"location"`
    CalendarRef string `yaml:"calendar_ref"`

    // propose_times
    IncludeFreebusy bool `yaml:"include_freebusy"`
    MaxCandidates   int  `yaml:"max_candidates"`

    // reschedule_event
    EventRef         string `yaml:"event_ref"`
    KeepParticipants bool   `yaml:"keep_participants"`

    // set/remove constraints
    Scope         string   `yaml:"scope"`
    ScopeRef      string   `yaml:"scope_ref"`
    ConstraintIDs []string `yaml:"constraint_ids"`

    // sync_now
    Direction string `yaml:"direction"` // pull|push|both
}

type ParticipantYAML struct {
    Email string `yaml:"email"`
    Role  string `yaml:"role"` // organizer|required|optional (default required)
}

type WindowYAML struct {
    Start string `yaml:"start"`
    End   string `yaml:"end"`
}

type ConstraintYAML struct {
    Kind    string                 `yaml:"kind"`
    Scope   string                 `yaml:"scope"`
    Payload map[string]interface{} `yaml:"payload"`
}

type SlotYAML struct {
    Start string `yaml:"start"`
    End   string `yaml:"end"`
}

type VoteYAML struct {
    SlotRef string `yaml:"slot_ref"` // slot id or ISO start
    Vote    string `yaml:"vote"`     // yes|no|maybe
    Email   string `yaml:"email"`    // optional – defaults to "unknown@example.com"
    Comment string `yaml:"comment,omitempty"`
}

func ParseDocument(r io.Reader) (*Document, []byte, error) {
    buf, err := io.ReadAll(r)
    if err != nil {
        return nil, nil, err
    }
    var doc Document
    if err := yaml.Unmarshal(buf, &doc); err != nil {
        return nil, nil, err
    }
    return &doc, buf, nil
}


