CREATE TABLE IF NOT EXISTS users(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    email varchar UNIQUE NOT NULL,
    password_hash varchar NOT NULL,
    first_name varchar,
    last_name varchar,
    active boolean NOT NULL DEFAULT true
);

CREATE INDEX idx_users_email ON users(email);
