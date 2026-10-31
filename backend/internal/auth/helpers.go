package auth

import (
	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
)

func updateStatusParams(id uuid.UUID, status int16) sqlcdb.UpdateUserStatusParams {
	return sqlcdb.UpdateUserStatusParams{ID: id, Status: status}
}
