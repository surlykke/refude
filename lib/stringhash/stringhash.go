package stringhash

const offset uint64 = 0xcbf29ce484222325 
const prime uint64 = 0x100000001b3 

func FNV1a(data...string) uint64 {
	var hash = offset
	for _, str := range data {
		for i := 0; i < len(str); i++ {
			hash = hash^uint64(str[i])
			hash = hash*prime
		}	
	}
	return hash
}
