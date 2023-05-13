package jamfactory

import "math/rand"

const jamLabelChars = "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

func (s *JamFactory) CreateLabel(depth int) string {
	if depth == 10 {
		panic("Recursion warning while creating a new label")
	}
	labelSlice := make([]byte, 5)
	for i := 0; i < 5; i++ {
		labelSlice[i] = jamLabelChars[rand.Intn(len(jamLabelChars))]
	}
	label := string(labelSlice)

	if ok, _ := s.JamLabels.Has(label); ok {
		return s.CreateLabel(depth + 1)
	}
	return label
}
