package spectrocloud

type Retry struct {
	runs          int
	retries       int
	expected_code int
}

type ResultStat struct {
	CODE_MINUS_ONE      int
	CODE_NORMAL         int
	CODE_EXPECTED       int
	CODE_INTERNAL_ERROR int
}
