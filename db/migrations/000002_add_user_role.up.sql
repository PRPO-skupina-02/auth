CREATE TYPE user_role AS ENUM ('customer', 'employee', 'admin');

ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'customer';

CREATE INDEX idx_users_role ON users(role);
