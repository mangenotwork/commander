package utils

// SliceIntersect 切片交集
func SliceIntersect(slice1 []string, slice2 []string) []string {
	m := make(map[string]int)
	n := make([]string,0)
	for _,v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times,_ := m[v]
		if times == 1 {
			n = append(n, v)
		}
	}
	return n
}

// SliceDifference 切片差集
func SliceDifference(slice1, slice2 []string) []string {
	m := make(map[string]int)
	n := make([]string, 0)
	inter := SliceIntersect(slice1, slice2)
	for _, v := range inter {
		m[v]=1
	}
	for _, v := range slice1 {
		if m[v] == 0 {
			n = append(n, v)
		}
	}
	for _, v := range slice2 {
		if m[v] == 0 {
			n = append(n, v)
		}
	}
	return n
}
