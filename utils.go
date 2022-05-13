package goeloquent

func CopyMap(original map[string]interface{}) (dest map[string]interface{}) {

	for k, v := range original {
		dest[k] = v
	}
	return
}
