package postgres

import (
	"context"
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func ExtractTxQuery(q *postgresDb.Queries, ctx context.Context) *postgresDb.Queries {
	tx, ok := GetTx(ctx)
	if ok {
		txQ := q.WithTx(tx)
		return txQ
	}

	return q
}

func MapPgNumericToDecimal(pgNumeric pgtype.Numeric) (decimal.Decimal, error) {
	if !pgNumeric.Valid {
		return decimal.Zero, fmt.Errorf("numeric value is invalid")
	}

	value, err := pgNumeric.Value()
	if err != nil {
		return decimal.Zero, err
	}

	sValue, ok := value.(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("failed to cast numeric value to string")
	}

	return decimal.NewFromString(sValue)
}

func MapDecimalToPgNumeric(decimal decimal.Decimal) (pgtype.Numeric, error) {
	pgNumeric := pgtype.Numeric{}

	err := pgNumeric.Scan(decimal.String())
	if err != nil {
		return pgNumeric, err
	}

	return pgNumeric, nil
}

func TimeToPgTimestampz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{
			Time:  t,
			Valid: false,
		}
	}
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func IntToPgInt2(n int16, isZeroValid bool) pgtype.Int2 {
	if n == 0 {
		return pgtype.Int2{
			Int16: n,
			Valid: isZeroValid,
		}
	}
	return pgtype.Int2{
		Int16: n,
		Valid: true,
	}
}
