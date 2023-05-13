package jamsession

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/permissions"
)

type Member struct {
	Identifier  string
	Permissions map[permissions.Permission]struct{}
}

func NewMember(userIdentifier string, p ...permissions.Permission) *Member {
	m := &Member{
		Identifier:  userIdentifier,
		Permissions: make(map[permissions.Permission]struct{}),
	}
	m.AddPermissions(p...)
	return m
}

func (m *Member) GetIdentifier() string {
	return m.Identifier
}

func (m *Member) GetPermissions() permissions.Permissions {
	p := make(permissions.Permissions, 0)
	for perm := range m.Permissions {
		p = append(p, perm)
	}
	return p
}

func (m *Member) SetPermissions(p ...permissions.Permission) {
	m.Permissions = make(map[permissions.Permission]struct{})
	m.AddPermissions(p...)
}

func (m *Member) AddPermissions(p ...permissions.Permission) {
	for _, toAdd := range p {
		m.Permissions[toAdd] = struct{}{}
	}
}

func (m *Member) RemovePermissions(p ...permissions.Permission) {
	for _, toRemove := range p {
		delete(m.Permissions, toRemove)
	}
}

func (m *Member) HasPermissions(p ...permissions.Permission) bool {
	for _, toCheck := range p {
		if _, ok := m.Permissions[toCheck]; !ok {
			return false
		}
	}
	return true
}

type Members map[string]*Member

func (m Members) Host() (*Member, error) {
	for _, member := range m {
		if member.HasPermissions(permissions.Host) {
			return member, nil
		}
	}
	return nil, errors.New("no host found")
}

func (m Members) Get(identifier string) (*Member, error) {
	if member, ok := m[identifier]; ok {
		return member, nil
	}
	return nil, errors.New("not a member")
}

func (m Members) Add(identifier string, p ...permissions.Permission) bool {
	if _, ok := m[identifier]; !ok {
		m[identifier] = NewMember(identifier, p...)
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
