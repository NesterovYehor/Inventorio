CREATE TABLE IF NOT EXISTS properties (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT 'Unnamed Property'
);

CREATE TABLE IF NOT EXISTS items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT DEFAULT 'Unnamed Item',
  quantity INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS property_needs (
  property_id INTEGER NOT NULL,
  item_id INTEGER NOT NULL,
  quantity INTEGER NOT NULL DEFAULT -1,
  PRIMARY KEY (property_id, item_id),
  FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE CASCADE,
  FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  is_draft BOOLEAN NOT NULL DEFAULT true,
  confirm_date DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS one_draft_order ON orders (is_draft)
WHERE
  is_draft = true;

CREATE TABLE IF NOT EXISTS arrivals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  property_id INTEGER NOT NULL,
  arrival_date TEXT NOT NULL,
  status TEXT DEFAULT 'pending',
  FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS order_lines (
  order_id INTEGER NOT NULL,
  item_id INTEGER NOT NULL,
  extra_qty INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (order_id, item_id),
  FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
  FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);
