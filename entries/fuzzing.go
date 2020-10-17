// +build gofuzz

package entries

// Fuzz is used by go-fuzz to fuzz the program.
func Fuzz(data []byte) int {
	p, err := NewParser("2020-06-01", "@!", "@?")
	if err != nil {
		return 0
	}

	_, err = p.Parse("test/entry", string(data))
	if err != nil {
		return 0
	}

	return 1
}
