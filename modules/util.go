package modules

func RemoveElement[T comparable](s *[]T, comp T) {
	res := []T{}
	for _, v:= range *s {
		if v != comp {
			res = append(res, v)
		}
	}
	*s = res
}
