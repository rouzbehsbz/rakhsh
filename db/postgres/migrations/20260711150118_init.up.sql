CREATE TABLE IF NOT EXISTS "clients" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    balance DECIMAL(36, 18) DEFAULT 0 NOT NULL,

    CONSTRAINT no_negative_balance CHECK (balance >= 0)
);