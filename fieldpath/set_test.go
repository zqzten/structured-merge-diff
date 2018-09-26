/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fieldpath

import (
	"testing"

	"github.com/kubernetes-sigs/structured-merge-diff/value"
)

func TestSetInsertHas(t *testing.T) {
	s1 := NewSet(
		MakePathOrDie("foo", 0, "bar", "baz"),
		MakePathOrDie("foo", 0, "bar"),
		MakePathOrDie("foo", 0),
		MakePathOrDie("foo", 1, "bar", "baz"),
		MakePathOrDie("foo", 1, "bar"),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("first")), "bar"),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("second")), "bar"),
	)

	table := []struct {
		set              *Set
		check            Path
		expectMembership bool
	}{
		{s1, MakePathOrDie("qux", KeyByFields("name", value.StringValue("second"))), false},
		{s1, MakePathOrDie("qux", KeyByFields("name", value.StringValue("second")), "bar"), true},
		{s1, MakePathOrDie("qux", KeyByFields("name", value.StringValue("first"))), true},
		{s1, MakePathOrDie("xuq", KeyByFields("name", value.StringValue("first"))), false},
		{s1, MakePathOrDie("foo", 0), true},
		{s1, MakePathOrDie("foo", 0, "bar"), true},
		{s1, MakePathOrDie("foo", 0, "bar", "baz"), true},
		{s1, MakePathOrDie("foo", 1), false},
		{s1, MakePathOrDie("foo", 1, "bar"), true},
		{s1, MakePathOrDie("foo", 1, "bar", "baz"), true},
	}

	for _, tt := range table {
		got := tt.set.Has(tt.check)
		if e, a := tt.expectMembership, got; e != a {
			t.Errorf("%v: wanted %v, got %v", tt.check.String(), e, a)
		}
	}
}

func TestSetEquals(t *testing.T) {
	table := []struct {
		a     *Set
		b     *Set
		equal bool
	}{
		{
			a:     NewSet(MakePathOrDie("foo")),
			b:     NewSet(MakePathOrDie("bar")),
			equal: false,
		},
		{
			a:     NewSet(MakePathOrDie("foo")),
			b:     NewSet(MakePathOrDie("foo")),
			equal: true,
		},
		{
			a:     NewSet(MakePathOrDie(1, "foo")),
			b:     NewSet(MakePathOrDie(0, "foo")),
			equal: false,
		},
		{
			a:     NewSet(MakePathOrDie(1, "foo")),
			b:     NewSet(MakePathOrDie(1, "foo", "bar")),
			equal: false,
		},
		{
			a: NewSet(
				MakePathOrDie(0),
				MakePathOrDie(1),
			),
			b: NewSet(
				MakePathOrDie(1),
				MakePathOrDie(0),
			),
			equal: true,
		},
		{
			a: NewSet(
				MakePathOrDie("foo", 0),
				MakePathOrDie("foo", 1),
			),
			b: NewSet(
				MakePathOrDie("foo", 1),
				MakePathOrDie("foo", 0),
			),
			equal: true,
		},
		{
			a: NewSet(
				MakePathOrDie("foo", 0),
				MakePathOrDie("foo"),
				MakePathOrDie("bar", "baz"),
				MakePathOrDie("qux", KeyByFields("name", value.StringValue("first"))),
			),
			b: NewSet(
				MakePathOrDie("foo", 1),
				MakePathOrDie("bar", "baz"),
				MakePathOrDie("bar"),
				MakePathOrDie("qux", KeyByFields("name", value.StringValue("second"))),
			),
			equal: false,
		},
	}

	for _, tt := range table {
		if e, a := tt.equal, tt.a.Equals(tt.b); e != a {
			t.Errorf("expected %v, got %v for:\na=\n%v\nb=\n%v", e, a, tt.a, tt.b)
		}
	}
}

func TestSetUnion(t *testing.T) {
	// Even though this is not a table driven test, since the thing under
	// test is recursive, we should be able to craft a single input that is
	// sufficient to check all code paths.

	s1 := NewSet(
		MakePathOrDie("foo", 0),
		MakePathOrDie("foo"),
		MakePathOrDie("bar", "baz"),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("first"))),
	)

	s2 := NewSet(
		MakePathOrDie("foo", 1),
		MakePathOrDie("bar", "baz"),
		MakePathOrDie("bar"),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("second"))),
	)

	u := NewSet(
		MakePathOrDie("foo", 0),
		MakePathOrDie("foo", 1),
		MakePathOrDie("foo"),
		MakePathOrDie("bar", "baz"),
		MakePathOrDie("bar"),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("qux", KeyByFields("name", value.StringValue("second"))),
	)

	got := s1.Union(s2)

	if !got.Equals(u) {
		t.Errorf("union: expected: \n%v\n, got \n%v\n", u, got)
	}
}

func TestSetIntersectionDifference(t *testing.T) {
	// Even though this is not a table driven test, since the thing under
	// test is recursive, we should be able to craft a single input that is
	// sufficient to check all code paths.

	s1 := NewSet(
		MakePathOrDie("a0"),
		MakePathOrDie("a1"),
		MakePathOrDie("foo", 0),
		MakePathOrDie("foo", 1),
		MakePathOrDie("b0", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("b1", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("bar", "c0"),
	)

	s2 := NewSet(
		MakePathOrDie("a1"),
		MakePathOrDie("a2"),
		MakePathOrDie("foo", 1),
		MakePathOrDie("foo", 2),
		MakePathOrDie("b1", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("b2", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("bar", "c2"),
	)

	i := NewSet(
		MakePathOrDie("a1"),
		MakePathOrDie("foo", 1),
		MakePathOrDie("b1", KeyByFields("name", value.StringValue("first"))),
	)

	got := s1.Intersection(s2)

	if !got.Equals(i) {
		t.Errorf("s1 intersect s2: expected: \n%v\n, got \n%v\n", i, got)
	}

	sDiffS2 := NewSet(
		MakePathOrDie("a0"),
		MakePathOrDie("foo", 0),
		MakePathOrDie("b0", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("bar", "c0"),
	)

	got = s1.Difference(s2)

	if !got.Equals(sDiffS2) {
		t.Errorf("s1 - s2: expected: \n%v\n, got \n%v\n", sDiffS2, got)
	}

	s2DiffS := NewSet(
		MakePathOrDie("a2"),
		MakePathOrDie("foo", 2),
		MakePathOrDie("b2", KeyByFields("name", value.StringValue("first"))),
		MakePathOrDie("bar", "c2"),
	)

	got = s2.Difference(s1)

	if !got.Equals(s2DiffS) {
		t.Errorf("s2 - s1: expected: \n%v\n, got \n%v\n", s2DiffS, got)
	}
}