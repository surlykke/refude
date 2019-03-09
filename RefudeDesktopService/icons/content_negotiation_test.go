package icons

import (
	"fmt"
	"testing"
)

func TestRead(t *testing.T) {
	fmt.Println(read("image/webp,image/apng,image/*,*/*;q=0.8,image/png;q=0.9", "application/json"))
}
