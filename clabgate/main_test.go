package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextReconcileIntervalBacksOffUntilMaximum(t *testing.T) {
	minimum := 30 * time.Second
	maximum := 2 * time.Minute

	assert.Equal(t, time.Minute, nextReconcileInterval(minimum, minimum, maximum, false))
	assert.Equal(t, maximum, nextReconcileInterval(time.Minute, minimum, maximum, false))
	assert.Equal(t, maximum, nextReconcileInterval(maximum, minimum, maximum, false))
}

func TestNextReconcileIntervalResetsAfterChange(t *testing.T) {
	assert.Equal(t, 30*time.Second, nextReconcileInterval(
		2*time.Minute,
		30*time.Second,
		2*time.Minute,
		true,
	))
}
