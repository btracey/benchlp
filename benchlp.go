/*
Copyright 2017 Brendan Tracey

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation and/or
other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may
be used to endorse or promote products derived from this software without specific
prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package benchlp

import "strconv"

type Term struct {
	Var   string
	Value float64
}

type Constraint struct {
	Left  []Term
	Right []Term
}

// WriteConstraints writes LP constraints as a string (would normally be written
// to a file).
//
// The constraints, in Go format, are represented as a set of left-hand-side terms
// and right-hand-side terms. For example
//  w1*v1 + w2*v6 <= w3*v1 + w4*v7
// would be represented as two terms on the left, and two terms on the right.
//
// A common LP file format requires that all of the variables be on the LHS and
// the constant term (ignored here) be on the right. For example,
//  (w1-w3)*v1 + w2*v5 - w4*v7 <=0
// WriteConstraints shifts the variables to one side, and converts the constraint
// to a []byte (with the real values for wi substituted).
func WriteConstraints(cons []Constraint, preallocate bool) {
	names, nameMap := IndexVariables(cons)

	// Temporary memory. constraintBytes overwrites and appends to b to reduce
	// allocations.
	var b []byte

	// NOTE(btracey): This is the hotspot. If these variables are pre-allocated,
	// then the GC does not run in the inner loop below, and a large chunck of
	// the running time is saved.
	var c1, c2 []float64
	if preallocate {
		c1 = make([]float64, len(names))
		c2 = make([]float64, len(names))
	}

	// Write constraints
	for _, c := range cons {
		b = b[:0]
		w := CondenseConstraint(c1, c2, c, nameMap)
		con := 0.0
		b = termBytes(b, w, names)
		b = append(b, []byte(" <= ")...)

		str := strconv.FormatFloat(con, 'g', 16, 64)
		b = append(b, []byte(str)...)
		b = append(b, []byte("\n")...)
	}
}

// termBytes appends all of the w_i * v_i terms.
func termBytes(b []byte, w []float64, names []string) []byte {
	first := true
	for i, v := range w {
		if v == 0 {
			continue
		}
		if !first {
			b = append(b, []byte(" + ")...)
		} else {
			first = false
		}
		str := strconv.FormatFloat(v, 'g', 16, 64)
		b = append(b, []byte(str)...)
		b = append(b, []byte(" ")...)
		b = append(b, []byte(names[i])...)
	}
	return b
}

// indexVariables assigns each variable to a unique index.
func IndexVariables(cons []Constraint) ([]string, map[string]int) {
	var names []string
	nameMap := make(map[string]int)

	for _, con := range cons {
		for _, term := range con.Left {
			names, nameMap = addNameIfNew(term.Var, names, nameMap)
		}
		for _, term := range con.Right {
			names, nameMap = addNameIfNew(term.Var, names, nameMap)
		}
	}
	return names, nameMap
}

// addNameIfNew adds the name to the list of names and the map if it is not already present.
func addNameIfNew(newName string, names []string, nameMap map[string]int) ([]string, map[string]int) {
	_, ok := nameMap[newName]
	if !ok {
		idx := len(names)
		names = append(names, newName)
		nameMap[newName] = idx
	}
	return names, nameMap
}

// CondenseTerms turns the slice of Term into a single weight vector where
// the value is for the variable with index i.
func CondenseTerms(w []float64, terms []Term, nameMap map[string]int) []float64 {
	nVar := len(nameMap)
	if w == nil {
		w = make([]float64, nVar)
	} else {
		for i := range w {
			w[i] = 0
		}
	}
	if len(w) != nVar {
		panic("lp: bad length")
	}
	for _, term := range terms {
		idx, ok := nameMap[term.Var]
		if !ok {
			panic("lp: term not present in name map")
		}
		w[idx] += term.Value
	}
	return w
}

// CondenseConstraints shifts all variables to the left hand side, and combines terms
// with the same variable.
func CondenseConstraint(wl, wr []float64, c Constraint, nameMap map[string]int) (w []float64) {
	wl = CondenseTerms(wl, c.Left, nameMap)
	wr = CondenseTerms(wr, c.Right, nameMap)

	sub(wl, wr) // move the terms to the left hand side
	return wl
}

// sub subtracts b from a
func sub(a, b []float64) {
	if len(a) != len(b) {
		panic("lp: slice length mismatch")
	}
	for i, v := range b {
		a[i] -= v
	}
}
