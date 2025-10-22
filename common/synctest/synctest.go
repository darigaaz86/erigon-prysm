// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package synctest

import (
	"context"
	"testing"
)

// Test is a helper for running tests with context support
func Test(t *testing.T, fn func(t *TestingT)) {
	t.Helper()
	ctx := context.Background()
	tt := &testingT{
		T:   t,
		ctx: ctx,
	}
	fn(tt)
}

// TestingT wraps testing.T with context support
type TestingT interface {
	testing.TB
	Context() context.Context
}

type testingT struct {
	*testing.T
	ctx context.Context
}

func (t *testingT) Context() context.Context {
	return t.ctx
}
