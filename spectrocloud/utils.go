package spectrocloud

func expandStringList(configured []interface{}) []string {
	vs := make([]string, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
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