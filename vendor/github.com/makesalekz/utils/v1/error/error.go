package error

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/lib/pq"
)

// Postgres Errors check library
// Source for Error Codes: https://www.postgresql.org/docs/current/errcodes-appendix.html

// ------- Constraint Violations -------
func IsNotNullViolation(err error) bool {
	var pqError *pq.Error
	ok := errors.As(err, &pqError)

	return ok && pqError.Code == "23502"
}

func IsForeignKeyViolation(err error) bool {
	var pqError *pq.Error
	ok := errors.As(err, &pqError)

	return ok && pqError.Code == "23503"
}

func IsUniqueViolation(err error) bool {
	var pqError *pq.Error
	ok := errors.As(err, &pqError)

	return ok && pqError.Code == "23505"
}
