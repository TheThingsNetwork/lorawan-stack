// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

type sfbw struct {
	sf uint32
	bw uint32
}

var demodulationFloor = map[sfbw]float32{
	{6, 125}:  -5,
	{7, 125}:  -7.5,
	{8, 125}:  -10,
	{9, 125}:  -12.5,
	{10, 125}: -15,
	{11, 125}: -17.5,
	{12, 125}: -20,

	// TODO: The values for BW250 and BW500 need to be verified
	{6, 250}:  -2,
	{7, 250}:  -4.5,
	{8, 250}:  -7,
	{9, 250}:  -9.5,
	{10, 250}: -12,
	{11, 250}: -14.5,
	{12, 250}: -17,

	{6, 500}:  1,
	{7, 500}:  -1.5,
	{8, 500}:  -4,
	{9, 500}:  -6.5,
	{10, 500}: -9,
	{11, 500}: -11.5,
	{12, 500}: -24,
}
