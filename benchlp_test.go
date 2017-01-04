package benchlp

import (
	"math/rand"
	"strconv"
	"testing"
)

func BenchmarkLPNoAllocate(b *testing.B) {
	benchmarkLP(b, false)
}

func BenchmarkLPAllocate(b *testing.B) {
	benchmarkLP(b, true)
}

func benchmarkLP(b *testing.B, preal bool) {
	nVars := 10000
	nConstraints := 50000
	cons := randomConstraints(nVars, nConstraints)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteConstraints(cons, preal)
	}
}

// randomConstraints generats a random set of sparse constraints.
func randomConstraints(nVars, nConstraints int) []Constraint {
	rnd := rand.New(rand.NewSource(0))
	var cons []Constraint
	for i := 0; i < nConstraints; i++ {
		con := Constraint{}

		nRightVars := int(rnd.ExpFloat64()) + 1
		for j := 0; j < nRightVars; j++ {
			idx := rnd.Intn(nVars)
			str := "v" + strconv.Itoa(idx)
			con.Left = append(con.Left, Term{str, rnd.Float64()})
		}

		nLeftVars := int(rnd.ExpFloat64()) + 1
		for j := 0; j < nLeftVars; j++ {
			idx := rnd.Intn(nVars)
			str := "v" + strconv.Itoa(idx)
			con.Right = append(con.Right, Term{str, rnd.Float64()})
		}
		cons = append(cons, con)
	}
	return cons
}

func BenchmarkAllocate(b *testing.B) {
	benchmarkNolp(b, true)
}

func BenchmarkAllocateNoCons(b *testing.B) {
	benchmarkNolp(b, false)
}

func benchmarkNolp(b *testing.B, withcons bool) {
	nVars := 10000
	nConstraints := 50000
	cons := []Constraint{
		{Left: []Term{{"a", 1}}},
	}
	if withcons {
		cons = randomConstraints(nVars, nConstraints)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c1 := make([]float64, nVars)
		c2 := make([]float64, nVars)
		_, _ = c1, c2
	}
	// Ensure cons stays alive
	v1 = cons[0].Left[0].Value
	v2 = cons[len(cons)-1].Left[0].Value
}

var v1, v2 float64 // So cons can't get GCd
