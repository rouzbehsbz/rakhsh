CREATE TABLE IF NOT EXISTS "clients" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    balance DECIMAL(36, 18) DEFAULT 0 NOT NULL,

    CONSTRAINT no_negative_balance CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS "messages" (
    uid BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    client_id INT NOT NULL,
    status SMALLINT NOT NULL,
    reason SMALLINT NULL,
    is_express BOOLEAN NOT NULL,
    recipient VARCHAR(20) NOT NULL,
    text VARCHAR(70) NOT NULL,

    PRIMARY KEY (uid)
) PARTITION BY RANGE (uid);