-- Таблица категорий
CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Таблица трат с внешним ключом на категорию
CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    amount      REAL    NOT NULL,
    description TEXT    NOT NULL,
    date        TEXT    NOT NULL,           -- ISO: YYYY-MM-DD
    category_id INTEGER,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- Базовый набор категорий
INSERT OR IGNORE INTO categories (name) VALUES
    ('Еда'),
    ('Транспорт'),
    ('Жильё'),
    ('Развлечения'),
    ('Здоровье'),
    ('Одежда'),
    ('Прочее');