package spectrocloud

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0)
	for _, v := range configured {
		if v != nil {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

func expandStringMap(configured map[string]interface{}) map[string]string {
	vs := make(map[string]string)
	for i, j := range configured {
		vs[i] = j.(string)
	}
	return vs
}

func stringContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func IsMapSubset[K, V comparable](m, sub map[K]V) bool {
	if len(sub) > len(m) {
		return false
	}
	for k, vsub := range sub {
		if vm, found := m[k]; !found || vm != vsub {
			return false
		}
	}
	return true
}
