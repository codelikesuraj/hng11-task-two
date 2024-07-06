package utils

import "testing"

func TestHashPassword(t *testing.T) {
	tests := []struct {
		password, hash string
		want           bool
	}{
		{"john", "$2y$14$aAU/ZeWti7PNj/OHT4y3XOnJ1MXsUkmAObRamRvC40FS.ISH2tt9q", true},
		{"jane", "$2y$14$bC8aQg7HWDktlutJusMct.xHQx5MU798lHN3tu6euNbX/XmelQhfe", true},
		{"doe", "$2y$14$tZUQ1MRsM7sxbzc0kqV1heROUi10DqqJW6O78lclzkuj63HnPGPfO", true},
		{"doe", "$2y$14$tZUQ1MRsM7sxbzc0kqV1heROUi10DqqJW6O78lclzkuj63HnPGPf1", false},
	}

	for _, test := range tests {
		if valid := PasswordIsValid(test.password, test.hash); valid != test.want {
			t.Errorf("wanted %t, got %t", test.want, valid)
		}
	}
}
