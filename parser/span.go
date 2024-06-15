// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package parser

import "fmt"

// A Span is a reference contiguous sequence of bytes in a query.
type Span struct {
	// Start is the index of the first byte of the span,
	// relative to the beginning of the query.
	Start int
	// End is the end index of the span (exclusive),
	// relative to the beginning of the query.
	End int
}

func newSpan(start, end int) Span {
	return Span{Start: start, End: end}
}

func indexSpan(i int) Span {
	return Span{Start: i, End: i}
}

func nullSpan() Span {
	return Span{Start: -1, End: -1}
}

// IsValid reports whether the span has a non-negative length
// and non-negative indices.
func (span Span) IsValid() bool {
	return span.Start >= 0 && span.End >= 0 && span.Start <= span.End
}

// Len returns the length of the span
// or zero if the span is invalid.
func (span Span) Len() int {
	if !span.IsValid() {
		return 0
	}
	return span.End - span.Start
}

// String formats the span indices as a mathematical range like "[12,34)".
func (span Span) String() string {
	return fmt.Sprintf("[%d,%d)", span.Start, span.End)
}

// Overlaps reports whether span and span2 intersect.
// In the case where both of the spans are zero-length,
// Overlaps reports true if the spans are equal and valid.
// In the case where only one of the spans is zero-length,
// Overlaps reports true if the zero-length span's start
// is between the other span's bounds, inclusive.
func (span Span) Overlaps(span2 Span) bool {
	if !span.IsValid() || !span2.IsValid() {
		return false
	}
	intersect := Span{
		Start: max(span.Start, span2.Start),
		End:   min(span.End, span2.End),
	}
	return intersect.IsValid()
}

func unionSpans(spans ...Span) Span {
	u := nullSpan()
	for _, span := range spans {
		if !span.IsValid() {
			continue
		}
		if u.IsValid() {
			u = newSpan(min(u.Start, span.Start), max(u.End, span.End))
		} else {
			u = span
		}
	}
	return u
}

func spanString(s string, span Span) string {
	if !span.IsValid() {
		return ""
	}
	return s[span.Start:span.End]
}
