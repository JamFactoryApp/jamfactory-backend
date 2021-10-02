package jamsession

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
)

type Members map[string]*Member

type Member struct {
	userIdentifier string
	rights         []types.MemberRights
}

func (m Members) Get(identifier string) (*Member, error) {
	if member, ok := m[identifier]; ok {
		return member, nil
	}
	return nil, errors.New("not a member")
}

func (m Members) Add(identifier string, rights []types.MemberRights) bool {
	if _, ok := m[identifier]; !ok {
		m[identifier] = &Member{
			userIdentifier: identifier,
			rights:         rights,
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

func (m *Member) Identifier() string {
	return m.userIdentifier
}

func (m *Member) Rights() []types.MemberRights {
	return m.rights
}

func (m *Member) SetRights(rights []types.MemberRights) {
	m.rights = rights
}

func (m *Member) HasRights(rights []types.MemberRights) bool {
	hasAllRights := true
	for _, wanted := range rights {
		if !ContainsRight(wanted, m.rights) {
			hasAllRights = false
		}
	}
	return hasAllRights
}

func (m *Member) AddRights(rights []types.MemberRights) error {
	for _, toAdd := range rights {
		if !ContainsRight(toAdd, m.rights) {
			m.rights = append(m.rights, toAdd)
		}
	}
	return nil
}

func (m *Member) RemoveRights(rights []types.MemberRights) {
	for _, toRemove := range rights {
		for i, right := range m.rights {
			if toRemove == right {
				m.rights = append(m.rights[:i], m.rights[i+1:]...)
				break
			}
		}
	}
}

func ContainsRight(right types.MemberRights, rights []types.MemberRights) bool {
	contains := false
	for _, r := range rights {
		if r == right {
			contains = true
		}
	}
	return contains
}

func ValidRights(rights []types.MemberRights) bool {
	for _, r := range rights {
		if !ContainsRight(r, types.ValidRights) {
			return false
		}
	}
	return true
}
