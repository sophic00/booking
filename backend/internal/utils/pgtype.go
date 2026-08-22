package utils

import (
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgtype converts uuid.UUID to pgtype.UUID
func UUIDToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// PgtypeToUUID converts pgtype.UUID to uuid.UUID
func PgtypeToUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return p.Bytes
}

// StringToPgtypeUUID parses a UUID string into pgtype.UUID
func StringToPgtypeUUID(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return UUIDToPgtype(u), nil
}

// TimeToPgtype converts time.Time to pgtype.Timestamptz
func TimeToPgtype(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

// PgtypeToTime converts pgtype.Timestamptz to time.Time
func PgtypeToTime(p pgtype.Timestamptz) time.Time {
	if !p.Valid {
		return time.Time{}
	}
	return p.Time
}

// Float64ToPgtypeNumeric converts float64 to pgtype.Numeric
func Float64ToPgtypeNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	str := fmt.Sprintf("%.2f", f)
	if err := num.Scan(str); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}

// PgtypeNumericToFloat64 converts pgtype.Numeric to float64
func PgtypeNumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0.0
	}
	if n.NaN || n.InfinityModifier != 0 {
		return 0.0
	}
	f, _ := n.Float64Value()
	if f.Valid {
		return f.Float64
	}

	// Fallback conversion from Int & Exp
	if n.Int != nil {
		bf := new(big.Float).SetInt(n.Int)
		if n.Exp < 0 {
			div := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil))
			bf.Quo(bf, div)
		} else if n.Exp > 0 {
			mult := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil))
			bf.Mul(bf, mult)
		}
		res, _ := bf.Float64()
		return res
	}

	return 0.0
}
