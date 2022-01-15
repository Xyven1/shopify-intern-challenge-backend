CREATE TABLE IF NOT EXISTS inventory (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  ammount INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS event_history (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  action TEXT NOT NULL,
  item_uid INTEGER NOT NULL,
  item_previous TEXT NOT NULL,
  timestamp datetime DEFAULT CURRENT_TIMESTAMP,
  comment TEXT NOT NULL
);