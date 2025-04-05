-- Create Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Talks table
CREATE TABLE IF NOT EXISTS talks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    speaker_id INTEGER NOT NULL,
    scheduled_date DATE,
    preferred_dates TEXT NOT NULL, -- JSON array of preferred dates
    status TEXT NOT NULL CHECK (status IN ('proposed', 'scheduled', 'completed', 'canceled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (speaker_id) REFERENCES users(id)
);

-- Create Votes table
CREATE TABLE IF NOT EXISTS votes (
    user_id INTEGER NOT NULL,
    talk_id INTEGER NOT NULL,
    interest_level INTEGER NOT NULL CHECK (interest_level BETWEEN 1 AND 5),
    availability TEXT NOT NULL, -- JSON object mapping dates to availability
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, talk_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (talk_id) REFERENCES talks(id)
);

-- Create Attendance table
CREATE TABLE IF NOT EXISTS attendance (
    talk_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('confirmed', 'attended', 'declined', 'no-show')),
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (talk_id, user_id),
    FOREIGN KEY (talk_id) REFERENCES talks(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create Resources table
CREATE TABLE IF NOT EXISTS resources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    talk_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('slides', 'video', 'code', 'article', 'other')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (talk_id) REFERENCES talks(id)
);

-- Create indices for foreign keys
CREATE INDEX IF NOT EXISTS idx_talks_speaker_id ON talks(speaker_id);
CREATE INDEX IF NOT EXISTS idx_votes_talk_id ON votes(talk_id);
CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id);
CREATE INDEX IF NOT EXISTS idx_attendance_talk_id ON attendance(talk_id);
CREATE INDEX IF NOT EXISTS idx_attendance_user_id ON attendance(user_id);
CREATE INDEX IF NOT EXISTS idx_resources_talk_id ON resources(talk_id);