// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package band

import (
	"fmt"
	"math"
	"slices"
)

// ChMaskCntlPair pairs a ChMaskCntl with a mask.
type ChMaskCntlPair struct {
	Cntl uint8
	Mask [16]bool
}

func parseChMask(offset uint8, mask ...bool) map[uint8]bool {
	if len(mask)-1 > int(math.MaxUint8-offset) {
		panic(fmt.Sprintf("channel mask overflows uint8, offset: %d, mask length: %d", offset, len(mask)))
	}
	m := make(map[uint8]bool, len(mask))
	for i, v := range mask {
		m[offset+uint8(i)] = v
	}
	return m
}

func parseChMask16(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0:
		return parseChMask(0, mask[:]...), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask48(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2:
		return parseChMask(cntl*16, mask[:]...), nil
	case 3:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	case 4:
		return parseChMask(0,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask64(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2, 3:
		return parseChMask(cntl*16, mask[:]...), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	case 7:
		return parseChMask(0,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask72(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2, 3:
		return parseChMask(cntl*16, mask[:]...), nil
	case 4:
		return parseChMask(64, mask[0:8]...), nil
	case 5:
		return parseChMask(0,
			mask[0], mask[0], mask[0], mask[0], mask[0], mask[0], mask[0], mask[0],
			mask[1], mask[1], mask[1], mask[1], mask[1], mask[1], mask[1], mask[1],
			mask[2], mask[2], mask[2], mask[2], mask[2], mask[2], mask[2], mask[2],
			mask[3], mask[3], mask[3], mask[3], mask[3], mask[3], mask[3], mask[3],
			mask[4], mask[4], mask[4], mask[4], mask[4], mask[4], mask[4], mask[4],
			mask[5], mask[5], mask[5], mask[5], mask[5], mask[5], mask[5], mask[5],
			mask[6], mask[6], mask[6], mask[6], mask[6], mask[6], mask[6], mask[6],
			mask[7], mask[7], mask[7], mask[7], mask[7], mask[7], mask[7], mask[7],
		), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			mask[0], mask[1], mask[2], mask[3], mask[4], mask[5], mask[6], mask[7],
		), nil
	case 7:
		return parseChMask(0,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false,
			mask[0], mask[1], mask[2], mask[3], mask[4], mask[5], mask[6], mask[7],
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func parseChMask96(mask [16]bool, cntl uint8) (map[uint8]bool, error) {
	switch cntl {
	case 0, 1, 2, 3, 4, 5:
		return parseChMask(cntl*16, mask[:]...), nil
	case 6:
		return parseChMask(0,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		), nil
	}
	return nil, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", cntl)
}

func boolsTo16BoolArray(vs ...bool) [16]bool {
	if len(vs) > 16 {
		panic(fmt.Sprintf("length of vs must be less or equal to 16, got %d", len(vs)))
	}
	var ret [16]bool
	for i, v := range vs {
		ret[i] = v
	}
	return ret
}

func generateChMask16(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 16 || len(desiredChs) != 16 {
		return nil, errInvalidChannelCount.New()
	}
	// NOTE: ChMaskCntl==6 never provides a more optimal ChMask sequence than ChMaskCntl==0.
	return []ChMaskCntlPair{
		{
			Mask: boolsTo16BoolArray(desiredChs...),
		},
	}, nil
}

// EqualChMasks returns true if both channel masks are equal.
func EqualChMasks(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateChMaskMatrix(pairs []ChMaskCntlPair, currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	n := len(currentChs)
	if n%16 != 0 || len(desiredChs) != n {
		return nil, errInvalidChannelCount.New()
	}
	for i := 0; i < n/16; i++ {
		for j := 0; j < 16; j++ {
			if currentChs[16*i+j] != desiredChs[16*i+j] {
				pairs = append(pairs, ChMaskCntlPair{
					Cntl: uint8(i),
					Mask: boolsTo16BoolArray(desiredChs[16*i : 16*i+16]...),
				})
				break
			}
		}
	}
	return pairs, nil
}

func trueCount(vs ...bool) int {
	var n int
	for _, v := range vs {
		if v {
			n++
		}
	}
	return n
}

func generateChMask48(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 48 || len(desiredChs) != 48 {
		return nil, errInvalidChannelCount.New()
	}
	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 3), currentChs[0:48], desiredChs[0:48])
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming either ChMaskCntl==3 or ChMaskCntl==4 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl3Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs[0:48])
	if err != nil {
		return nil, err
	}
	cntl4Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
	}, desiredChs[0:48])
	if err != nil {
		return nil, err
	}
	switch {
	case len(pairs) <= 1+len(cntl3Pairs) && len(pairs) <= 1+len(cntl4Pairs):
		return pairs, nil

	case len(cntl3Pairs) < len(cntl4Pairs):
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl3Pairs)), ChMaskCntlPair{
			Cntl: 3,
		}), cntl3Pairs...), nil

	default:
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl4Pairs)), ChMaskCntlPair{
			Cntl: 4,
		}), cntl4Pairs...), nil
	}
}

func generateChMask64(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 64 || len(desiredChs) != 64 {
		return nil, errInvalidChannelCount.New()
	}
	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 5), currentChs[0:64], desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming either ChMaskCntl==6 or ChMaskCntl==7 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl6Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	cntl7Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	switch {
	case len(pairs) <= 1+len(cntl6Pairs) && len(pairs) <= 1+len(cntl7Pairs):
		return pairs, nil

	case len(cntl6Pairs) < len(cntl7Pairs):
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl6Pairs)), ChMaskCntlPair{
			Cntl: 6,
		}), cntl6Pairs...), nil

	default:
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl7Pairs)), ChMaskCntlPair{
			Cntl: 7,
		}), cntl7Pairs...), nil
	}
}

func generateChMask72Generic(currentChs, desiredChs []bool, atomic bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 72 || len(desiredChs) != 72 {
		return nil, errInvalidChannelCount.New()
	}
	if EqualChMasks(currentChs, desiredChs) {
		return []ChMaskCntlPair{
			{
				Mask: boolsTo16BoolArray(desiredChs[0:16]...),
			},
		}, nil
	}

	on125 := trueCount(desiredChs[0:64]...)
	switch on125 {
	case 0:
		return []ChMaskCntlPair{
			{
				Cntl: 7,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			},
		}, nil

	case 64:
		return []ChMaskCntlPair{
			{
				Cntl: 6,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			},
		}, nil
	}

	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 5), currentChs[0:64], desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	for i := 65; i < 72; i++ {
		if currentChs[i] != desiredChs[i] {
			pairs = append(pairs, ChMaskCntlPair{
				Cntl: 4,
				Mask: boolsTo16BoolArray(desiredChs[64:72]...),
			})
			break
		}
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming either ChMaskCntl==6 or ChMaskCntl==7 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl6Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	cntl7Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 4), []bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
	}, desiredChs[0:64])
	if err != nil {
		return nil, err
	}
	switch {
	case len(pairs) <= 1+len(cntl6Pairs) && len(pairs) <= 1+len(cntl7Pairs):
		return pairs, nil

	case len(cntl6Pairs) < len(cntl7Pairs):
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl6Pairs)), ChMaskCntlPair{
			Cntl: 6,
			Mask: boolsTo16BoolArray(desiredChs[64:72]...),
		}), cntl6Pairs...), nil

	case !atomic:
		// If the masks are not atomically processed, the Cntl=7 command will appear to the
		// end device as an attempt to completely mute the end device, and it will be rejected.
		// In such situations, we will fall back to including all of the pairs.
		// We also sort the masks descending on the number of enabled channels in order to ensure
		// that intermediary states through which the end device goes while parsing the masks
		// are valid (i.e. the total number of enabled channels is always greater than 0).
		slices.SortFunc(pairs, func(a, b ChMaskCntlPair) int {
			return trueCount(b.Mask[:]...) - trueCount(a.Mask[:]...)
		})
		return pairs, nil

	default:
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl7Pairs)), ChMaskCntlPair{
			Cntl: 7,
			Mask: boolsTo16BoolArray(desiredChs[64:72]...),
		}), cntl7Pairs...), nil
	}
}

func makeGenerateChMask72(supportChMaskCntl5 bool, atomic bool) func([]bool, []bool) ([]ChMaskCntlPair, error) {
	if !supportChMaskCntl5 {
		return func(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
			return generateChMask72Generic(currentChs, desiredChs, atomic)
		}
	}
	return func(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
		pairs, err := generateChMask72Generic(currentChs, desiredChs, atomic)
		if err != nil {
			return nil, err
		}
		if len(pairs) <= 1 {
			return pairs, nil
		}

		var fsbs [8]bool
		for i := 0; i < 8; i++ {
			if trueCount(desiredChs[8*i:8*i+8]...) == 8 {
				fsbs[i] = true
			}
		}
		if n := trueCount(fsbs[:]...); n == 0 || n == 8 {
			// Since there are either no enabled FSBs, or no disabled FSBs we won't be able to compute a
			// more efficient result that one using ChMaskCntl==6 or ChMaskCntl==7.
			return pairs, nil
		}
		cntl5Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 5), []bool{
			fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0], fsbs[0],
			fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1], fsbs[1],
			fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2], fsbs[2],
			fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3], fsbs[3],
			fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4], fsbs[4],
			fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5], fsbs[5],
			fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6], fsbs[6],
			fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7], fsbs[7],
		}, desiredChs[0:64])
		if err != nil {
			return nil, err
		}
		for i := 65; i < 72; i++ {
			if currentChs[i] != desiredChs[i] {
				cntl5Pairs = append(cntl5Pairs, ChMaskCntlPair{
					Cntl: 4,
					Mask: boolsTo16BoolArray(desiredChs[64:72]...),
				})
				break
			}
		}
		if len(pairs) <= 1+len(cntl5Pairs) {
			return pairs, nil
		}
		return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl5Pairs)), ChMaskCntlPair{
			Cntl: 5,
			Mask: boolsTo16BoolArray(fsbs[:]...),
		}), cntl5Pairs...), nil
	}
}

func generateChMask96(currentChs, desiredChs []bool) ([]ChMaskCntlPair, error) {
	if len(currentChs) != 96 || len(desiredChs) != 96 {
		return nil, errInvalidChannelCount.New()
	}
	if EqualChMasks(currentChs, desiredChs) {
		return []ChMaskCntlPair{
			{
				Mask: boolsTo16BoolArray(desiredChs[0:16]...),
			},
		}, nil
	}
	if trueCount(desiredChs...) == 96 {
		return []ChMaskCntlPair{
			{
				Cntl: 6,
			},
		}, nil
	}
	pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 6), currentChs, desiredChs)
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 2 {
		return pairs, nil
	}
	// Count amount of pairs required assuming ChMaskCntl==6 is sent first.
	// The minimum amount of pairs required in such case will be 2, hence only attempt this if amount
	// of generated pairs so far is higher than 2.
	cntl6Pairs, err := generateChMaskMatrix(make([]ChMaskCntlPair, 0, 6), []bool{
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}, desiredChs)
	if err != nil {
		return nil, err
	}
	if len(pairs) <= 1+len(cntl6Pairs) {
		return pairs, nil
	}
	return append(append(make([]ChMaskCntlPair, 0, 1+len(cntl6Pairs)), ChMaskCntlPair{
		Cntl: 6,
	}), cntl6Pairs...), nil
}
