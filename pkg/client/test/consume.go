package test

import "testing"

func consumeResults(t *testing.T, retry Retry, ch chan int, done chan bool) ResultStat {
	stat := ResultStat{
		CODE_MINUS_ONE:      0,
		CODE_NORMAL:         0,
		CODE_EXPECTED:       0,
		CODE_INTERNAL_ERROR: 0,
	}

	for i := 0; i < retry.runs; i++ {
		v := <-ch
		switch v {
		case -1:
			stat.CODE_MINUS_ONE++
			break
		case retry.expected_code:
			stat.CODE_EXPECTED++
			break
		case 200:
			stat.CODE_NORMAL++
			break
		case 500:
			stat.CODE_INTERNAL_ERROR++
			break
		default:
			t.Fail()
		}
	}
	<-done
	return stat
}
