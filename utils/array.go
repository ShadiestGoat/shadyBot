package utils

func NoDupeAppend(curArr, newItems []string) []string {
	m := map[string]bool{}

	for _, c := range curArr {
		m[c] = true
	}

	for _, n := range newItems {
		if m[n] {
			continue
		}
		curArr = append(curArr, n)
	}

	return curArr
}

// Returns -1 if the searchTerm is not found
func BinarySearch[T int | string](a []T, searchTerm T) (location int) {
	mid := len(a) / 2

	switch {
	case len(a) == 0:
		location = -1 // not found
	case a[mid] > searchTerm:
		location = BinarySearch(a[:mid], searchTerm)
	case a[mid] < searchTerm:
		location = BinarySearch(a[mid+1:], searchTerm)
		if location >= 0 { // if anything but the -1 "not found" result
			location += mid + 1
		}
	default: // a[mid] == search
		location = mid // found
	}

	return
}
