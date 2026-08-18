package utils 

func Map[F,T any](s []F, mapper func(f F) T) []T{
	var res = make([]T, len(s))
	for i, f := range s {
		res[i] = mapper(f)
	}
	return res
}
