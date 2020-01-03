package numbers

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMAE(t *testing.T) {
	a := NewMAE(10000)

	for i := 0; i < 100000; i++ {
		r1 := rand.Float64() * 100000
		r2 := rand.Float64() * 100000
		a.Apply(r1, r2)
		print(r1, r2)
		assert.True(t, a.Get() >= 0)
		println(" ==> ", a.Get())
	}
}

func TestMAENegative(t *testing.T) {
	a := NewMAE(10000)
	a.Apply(60466.03, 94050.91)
	a.Apply(66456.01, 43771.42)
	a.Apply(42463.75, 68682.31)
	a.Apply(6563.702, 15651.93)
	a.Apply(9696.952, 30091.19)
	a.Apply(51521.26, 81364.00)
	a.Apply(21426.39, 38065.72)
	a.Apply(31805.82, 46888.98)
	a.Apply(28303.42, 29310.19)
	assert.GreaterOrEqual(t, a.Get(), 0.0)
}

func TestABSDIFFNegative(t *testing.T) {
	assert.GreaterOrEqual(t, ABSDIFF(60466.03, 94050.91), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(66456.01, 43771.42), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(42463.75, 68682.31), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(6563.702, 15651.93), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(9696.952, 30091.19), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(51521.26, 81364.00), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(21426.39, 38065.72), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(31805.82, 46888.98), 0.0)
	assert.GreaterOrEqual(t, ABSDIFF(28303.42, 29310.19), 0.0)
}
