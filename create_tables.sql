CREATE TABLE IF NOT EXISTS Inventory (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  ammount INTEGER DEFAULT 0 NOT NULL
);