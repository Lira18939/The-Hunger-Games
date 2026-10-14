CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

INSERT OR IGNORE INTO categories (id, name) VALUES
    (1, 'Еда'),
    (2, 'Транспорт'),
    (3, 'Жильё'),
    (4, 'Развлечения'),
    (5, 'Прочее');

ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id);