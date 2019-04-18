// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package requests

import (
	"fmt"
	"testing"
)

func TestRead(t *testing.T) {
	fmt.Println(read("image/webp,image/apng,image/*,*/*;q=0.8,image/png;q=0.9", "application/json"))
}
