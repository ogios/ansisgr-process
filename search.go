package process

type BoundsStruct interface {
	getBounds() [2]int
}

// binary search index given `Bound [2]int`
func Search(list []BoundsStruct, pos int) []int {
	listLen := len(list)
	if listLen == 0 {
		return []int{-1, -1}
	}
	i := (listLen - 1) / 2
	step := i
	halfStep := func() {
		step /= 2
		if step == 0 {
			step = 1
		}
	}
	// 1. index between one bounds start & end
	// or
	// 2.between ( last end and current start ) or ( current end and next start )
	for {
		halfStep()
		v := list[i]
		vBound := v.getBounds()
		// between bounds
		// or, width is 0 and index is the same
		if vBound[0] <= pos && vBound[1] > pos {
			return []int{i}
		} else {
			// smaller than start
			if pos < vBound[0] {
				if i > 0 {
					// not the first one
					prev := list[i-1]
					if prev.getBounds()[1] <= pos {
						// i bigger than prev end and i smaller than current start means circumstance 2
						return []int{i - 1, i}
					} else {
						// i smaller than prev end means still space to go left
						i -= step
					}
				} else {
					// first one and i smaller than first start means circumstance 2
					return []int{-1, i}
				}
			} else if pos >= vBound[1] {
				// bigger than end
				if i < listLen-1 {
					// not the last one
					next := list[i+1]
					if pos < next.getBounds()[0] {
						// i bigger than current end and smaller than next start means circumstance 2
						return []int{i, i + 1}
					} else {
						// i bigger or equal to next start means still space to go right
						i += step
					}
				} else {
					// last one and i bigger than end means circumstance 2
					// return []int{i, i + 1}
					return []int{i, -1}
				}
			}
		}
	}
}
