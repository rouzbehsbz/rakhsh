-- name: InsertMessage :exec
INSERT INTO "messages" (
    uid, 
    created_at, 
    updated_at, 
    client_id, 
    status, 
    reason, 
    is_express, 
    recipient, 
    text
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: UpdateMessage :exec
UPDATE messages
SET
    status = $1,
    reason = $2,
    updated_at = NOW()
WHERE uid = $3;

-- name: BatchUpdateMessages :exec
UPDATE messages AS m
SET
    status     = u.new_status,
    reason     = u.new_reason,
    updated_at = NOW()
FROM (
    SELECT 
        UNNEST($1::bigint[])   AS target_uid,
        UNNEST($2::smallint[]) AS new_status,
        UNNEST($3::smallint[]) AS new_reason
) AS u
WHERE m.uid = u.target_uid;

-- name: FindMessageByUid :one
SELECT * FROM "messages" WHERE client_id = $1 AND uid = $2;

-- name: FindAllMessagesByUids :many
SELECT * FROM messages WHERE client_id = $1 AND uid = ANY($2::bigint[]);