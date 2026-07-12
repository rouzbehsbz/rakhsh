-- Message Status Enum
-- 0: Pending, 1: Submitted, 2: Delivered, 3: Rejected

-- Message Reason Enum
-- 0: Internal Error, 1: Operator Error, 2: 

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