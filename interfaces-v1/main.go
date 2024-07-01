/*
Input: position = [1,2,3,4,7], m = 3
Output: 3
*/

package main

import (
	"sort"
)

func maxDistance(position []int, m int) int {
	_ = m
	pos := sort.IntSlice(position)
	return pos[0]
	/* for i, v := range pos {
		if i < max(i, pos) {

		}

	} */
}

func main() {
	_ = maxDistance([]int{}, 3)
}
