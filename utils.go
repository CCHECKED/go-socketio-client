package socketio

func Contains[T comparable](slice []T, str T) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
