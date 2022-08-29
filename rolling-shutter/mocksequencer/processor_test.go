package mocksequencer

import (
	"testing"

	"gotest.tools/assert"
)

func assertFind[T any](t *testing.T, mp *activationBlockMap[T], block int, expectedValue T) {
	t.Helper()
	found, err := mp.Find(uint64(block))
	assert.NilError(t, err)
	assert.Equal(t, found, expectedValue)
}

func assertError[T any](t *testing.T, mp *activationBlockMap[T], block int, errMessage string) {
	t.Helper()
	_, err := mp.Find(uint64(block))
	assert.Error(t, err, errMessage)
}

func TestActivationBlockMap(t *testing.T) {
	mp := newActivationBlockMap[string]()
	mp.Set("third", 5)
	mp.Set("first", 0)
	mp.Set("second", 1)

	assertFind(t, mp, 0, "first")
	assertFind(t, mp, 1, "second")
	assertFind(t, mp, 4, "second")
	assertFind(t, mp, 5, "third")
	assertFind(t, mp, 6, "third")
	assertFind(t, mp, 999999, "third")
}

func TestActivationBlockMapErrors(t *testing.T) {
	mp := newActivationBlockMap[uint64]()
	assertError(t, mp, 0, "no value was set")

	mp.Set(42, 10)
	assertError(t, mp, 9, "no value was found")
}
