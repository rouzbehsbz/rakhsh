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