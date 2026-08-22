package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDConversions(t *testing.T) {
	originalUUID := uuid.New()
	pgUUID := UUIDToPgtype(originalUUID)
	assert.True(t, pgUUID.Valid)

	convertedBack := PgtypeToUUID(pgUUID)
	assert.Equal(t, originalUUID, convertedBack)

	parsed, err := StringToPgtypeUUID(originalUUID.String())
	assert.NoError(t, err)
	assert.Equal(t, originalUUID, PgtypeToUUID(parsed))

	_, err = StringToPgtypeUUID("invalid-uuid")
	assert.Error(t, err)
}

func TestTimeConversions(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	pgTime := TimeToPgtype(now)
	assert.True(t, pgTime.Valid)

	convertedBack := PgtypeToTime(pgTime)
	assert.True(t, now.Equal(convertedBack))
}

func TestNumericConversions(t *testing.T) {
	amount := 25.50
	pgNum := Float64ToPgtypeNumeric(amount)
	assert.True(t, pgNum.Valid)

	convertedBack := PgtypeNumericToFloat64(pgNum)
	assert.InDelta(t, amount, convertedBack, 0.001)
}
