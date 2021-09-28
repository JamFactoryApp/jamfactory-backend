package jamsession

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
)

type Members map[string]Member

type Member struct {
	UserIdentifier string
	Rights         []types.MemberRights
}

func (m Members) Get(identifier string) (*Member, error) {
	if member, ok := m[identifier]; ok {
		return &member, nil
	}
	return nil, errors.New("not a member")
}

func (m Members) Add(identifier string, rights []types.MemberRights) bool {
	if _, ok := m[identifier]; !ok {
		m[identifier] = Member{
			UserIdentifier: identifier,
			Rights:         rights,
		}
		return true
	}
	return false
}

func (m Members) Remove(identifier string) bool {
	if _, ok := m[identifier]; ok {
		delete(m, identifier)
		return true
	}
	return false
}

func (m *Member) Has(rights []types.MemberRights) bool {
	hasAllRights := true
	for _, wanted := range rights {
		hasCurrent := false
		for _, right := range m.Rights {
			if right == wanted {
				hasCurrent = true
				break
			}
		}
		if !hasCurrent {
			hasAllRights = false
		}
	}
	return hasAllRights
}

func (m *Member) Add(rights []types.MemberRights) {
	for _, toAdd := range rights {
		if !m.Has([]types.MemberRights{toAdd}) {
			m.Rights = append(m.Rights, toAdd)
		}
	}
}

func (m *Member) Remove(rights []types.MemberRights) bool {
	removedAllRights := true
	for _, toRemove := range rights {
		removedCurrent := false
		for i, right := range m.Rights {
			if toRemove == right {
				m.Rights = append(m.Rights[:i], m.Rights[i+1:]...)
				removedCurrent = true
				break
			}
		}
		if !removedCurrent {
			removedAllRights = false
		}
	}
	return removedAllRights
}
