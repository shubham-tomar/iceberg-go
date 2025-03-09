// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package internal

import (
	"cmp"
	"slices"
)

// Helper function to find the difference between two slices (a - b).
func Difference(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range b {
		m[item] = true
	}

	diff := make([]string, 0)
	for _, item := range a {
		if !m[item] {
			diff = append(diff, item)
		}
	}

	return diff
}

type Bin[T any] struct {
	binWeight    int64
	targetWeight int64
	items        []T
}

func (b *Bin[T]) Weight() int64            { return b.binWeight }
func (b *Bin[T]) CanAdd(weight int64) bool { return b.binWeight+weight <= b.targetWeight }
func (b *Bin[T]) Add(item T, weight int64) {
	b.binWeight += weight
	b.items = append(b.items, item)
}

type SlicePacker[T any] struct {
	TargetWeight    int64
	Lookback        int
	LargestBinFirst bool
}

func (s *SlicePacker[T]) Pack(items []T, weightFunc func(T) int64) [][]T {
	bins := make([]Bin[T], 0)
	findBin := func(weight int64) *Bin[T] {
		for i := range bins {
			if bins[i].CanAdd(weight) {
				return &bins[i]
			}
		}

		return nil
	}

	removeBin := func() Bin[T] {
		if s.LargestBinFirst {
			maxBin := slices.MaxFunc(bins, func(a, b Bin[T]) int {
				return cmp.Compare(a.Weight(), b.Weight())
			})
			i := slices.IndexFunc(bins, func(e Bin[T]) bool {
				return e.Weight() == maxBin.Weight()
			})

			bins = slices.Delete(bins, i, i+1)

			return maxBin
		}

		var out Bin[T]
		out, bins = bins[0], bins[1:]

		return out
	}

	return slices.Collect(func(yield func([]T) bool) {
		for _, item := range items {
			w := weightFunc(item)
			bin := findBin(w)
			if bin != nil {
				bin.Add(item, w)
			} else {
				bin := Bin[T]{targetWeight: s.TargetWeight}
				bin.Add(item, w)
				bins = append(bins, bin)

				if len(bins) > s.Lookback {
					if !yield(removeBin().items) {
						return
					}
				}
			}
		}

		for len(bins) > 0 {
			if !yield(removeBin().items) {
				return
			}
		}
	})
}

func (s *SlicePacker[T]) PackEnd(items []T, weightFunc func(T) int64) [][]T {
	slices.Reverse(items)
	packed := s.Pack(items, weightFunc)
	slices.Reverse(packed)

	result := make([][]T, 0, len(packed))
	for _, items := range packed {
		slices.Reverse(items)
		result = append(result, items)
	}

	return result
}
