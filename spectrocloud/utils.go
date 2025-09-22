package spectrocloud

import (
	"math"
	"time"
)

// SafeInt32 converts int to int32 with bounds checking to prevent overflow
func SafeInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

// SafeInt64 converts int to int64 with bounds checking to prevent overflow
func SafeInt64(value int) int64 {

	return int64(value)
}

// SafeUint32 converts int to uint32 with bounds checking to prevent overflow
func SafeUint32(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxInt {
		return math.MaxUint32
	}
	return uint32(value)
}

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

func Bool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func BoolPtr(b bool) *bool {
	bol := b
	return &bol
}

func Int(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func IntPtr(i int) *int {
	it := i
	return &it
}

func Int8(i *int8) int8 {
	if i == nil {
		return 0
	}
	return *i
}

func Int8WithDefault(i *int8, defaultVal int8) int8 {
	if i == nil {
		return defaultVal
	}
	return *i
}

func Int8Ptr(i int8) *int8 {
	it := i
	return &it
}

func Int16(i *int16) int16 {
	if i == nil {
		return 0
	}
	return *i
}

func Int16WithDefault(i *int16, defaultVal int16) int16 {
	if i == nil {
		return defaultVal
	}
	return *i
}

func Int16Ptr(i int16) *int16 {
	it := i
	return &it
}

func Int32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func Int32WithDefault(i *int32, defaultVal int32) int32 {
	if i == nil {
		return defaultVal
	}
	return *i
}

func Int32Ptr(i int32) *int32 {
	it := i
	return &it
}

func Int64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func Int64WithDefault(i *int64, defaultVal int64) int64 {
	if i == nil {
		return defaultVal
	}
	return *i
}

func Int64Ptr(i int64) *int64 {
	it := i
	return &it
}

func Float32Ptr(f float32) *float32 {
	it := f
	return &it
}

func Float32WithDefault(f *float32, defaultVal float32) float32 {
	if f == nil {
		return defaultVal
	}
	return *f
}

func Float64Ptr(f float64) *float64 {
	it := f
	return &it
}

func Float64WithDefault(f *float64, defaultVal float64) float64 {
	if f == nil {
		return defaultVal
	}
	return *f
}

func String(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringWithDefaultValue(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

func StringPtr(s string) *string {
	str := s
	return &str
}

func Time(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func Interface(val *interface{}) interface{} {
	if val == nil {
		return nil
	}
	return val
}
