-- name: CreateDataset :one
INSERT INTO baseline_datasets (name, version, source_name, source_url, source_year, description,
    methodology, disclaimer, status, active, is_demo, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, FALSE, $10, $11)
RETURNING *;

-- name: GetDatasetByID :one
SELECT * FROM baseline_datasets WHERE id = $1;

-- name: GetActiveDataset :one
SELECT * FROM baseline_datasets WHERE active LIMIT 1;

-- name: ListDatasets :many
SELECT * FROM baseline_datasets ORDER BY created_at DESC;

-- name: DeactivateAllDatasets :exec
UPDATE baseline_datasets SET active = FALSE WHERE active;

-- name: SetDatasetActive :exec
UPDATE baseline_datasets SET active = TRUE, status = 2 WHERE id = $1;

-- name: UpdateDatasetStatus :exec
UPDATE baseline_datasets SET status = $2 WHERE id = $1;

-- name: UpdateDatasetStats :exec
UPDATE baseline_datasets SET record_count = $2, status = $3 WHERE id = $1;

-- name: DeleteBaselineRecords :exec
DELETE FROM baseline_records WHERE dataset_id = $1;

-- name: CountBaselineRecords :one
SELECT COUNT(*) FROM baseline_records WHERE dataset_id = $1;

-- name: CreateCalculatorRun :one
INSERT INTO calculator_runs (user_id, anonymous_session_id, dataset_id, input_snapshot, result_snapshot, share_token)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCalculatorRunByShareToken :one
SELECT * FROM calculator_runs WHERE share_token = $1;

-- name: UpdateCalculatorRunShareToken :exec
UPDATE calculator_runs SET share_token = $2 WHERE id = $1;

-- name: CountCalculatorRuns :one
SELECT COUNT(*) FROM calculator_runs;
