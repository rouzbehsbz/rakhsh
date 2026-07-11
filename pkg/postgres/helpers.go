package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

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
