package jamlabel

import (
	"math/rand"
)

type DefaultManager struct {
	jamLabels map[string]bool
}

const jamLabelChars = "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

func NewDefault() *DefaultManager {
	return &DefaultManager{
		jamLabels: make(map[string]bool),
	}
}

func (s *DefaultManager) List() []string {
	labels := make([]string, len(s.jamLabels))
	i := 0
	for j := range s.jamLabels {
		labels[i] = j
		i++
	}
	return labels
}

func (s *DefaultManager) Create() string {
	labelSlice := make([]byte, 5)
	for i := 0; i < 5; i++ {
		labelSlice[i] = jamLabelChars[rand.Intn(len(jamLabelChars))]
	}
	label := string(labelSlice)

	if _, exists := s.jamLabels[label]; exists {
		return s.Create()
	}

	s.jamLabels[label] = true
	return label
}

func (s *DefaultManager) Delete(jamLabel string) error {
	if _, exists := s.jamLabels[jamLabel]; !exists {
		return ErrJamLabelNotFound
	}
	s.jamLabels[jamLabel] = false
	return nil
}
