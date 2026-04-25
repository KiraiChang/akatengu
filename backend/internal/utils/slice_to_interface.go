package utils

// Int64SliceToInterface 將 []int64 轉成 []interface{} 方便 sqlx.In 使用
func Int64SliceToInterface(slice []int64) []interface{} {
	res := make([]interface{}, len(slice))
	for i, v := range slice {
		res[i] = v
	}
	return res
}

// StringSliceToInterface 將 []string 轉成 []interface{} 方便 sqlx.In 使用
func StringSliceToInterface(slice []string) []interface{} {
	res := make([]interface{}, len(slice))
	for i, v := range slice {
		res[i] = v
	}
	return res
}
