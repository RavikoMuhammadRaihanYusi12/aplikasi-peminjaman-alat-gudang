CREATE TABLE IF NOT EXISTS technicians (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT    NOT NULL,
  section     TEXT    NOT NULL DEFAULT '',
  active      INTEGER NOT NULL DEFAULT 1,
  created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS tools (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  code           TEXT    NOT NULL UNIQUE,
  name           TEXT    NOT NULL,
  category       TEXT    NOT NULL DEFAULT '',
  max_hours      INTEGER,
  condition      TEXT    NOT NULL DEFAULT 'baik' CHECK (condition IN ('baik','rusak')),
  condition_note TEXT    NOT NULL DEFAULT '',
  active         INTEGER NOT NULL DEFAULT 1,
  created_at     TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS loans (
  id               INTEGER PRIMARY KEY AUTOINCREMENT,
  tool_id          INTEGER NOT NULL REFERENCES tools(id),
  technician_id    INTEGER NOT NULL REFERENCES technicians(id),
  borrowed_at      TEXT    NOT NULL,
  borrow_shift     TEXT    NOT NULL,
  returned_at      TEXT,
  return_shift     TEXT,
  return_condition TEXT    CHECK (return_condition IN ('baik','rusak')),
  return_note      TEXT    NOT NULL DEFAULT '',
  voided_at        TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_loans_one_open_per_tool
  ON loans(tool_id) WHERE returned_at IS NULL AND voided_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_loans_tool_history ON loans(tool_id, borrowed_at DESC);
CREATE INDEX IF NOT EXISTS ix_loans_technician   ON loans(technician_id);

CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

INSERT OR IGNORE INTO settings (key, value) VALUES ('default_max_hours', '24');
INSERT OR IGNORE INTO settings (key, value) VALUES ('shift_pagi_start', '07:00');
INSERT OR IGNORE INTO settings (key, value) VALUES ('shift_malam_start', '19:00');

