package jamsession

import (
	"errors"

	"github.com/jamfactoryapp/jamfactory-backend/api/types"
)

type Members map[string]*Member

type Member struct {
	userIdentifier string
	permissions    []types.Permission
}

func (m Members) Get(identifier string) (*Member, error) {
	if member, ok := m[identifier]; ok {
		return member, nil
	}
	return nil, errors.New("not a member")
}

func (m Members) Add(identifier string, permissions []types.Permission) bool {
	if _, ok := m[identifier]; !ok {
		m[identifier] = &Member{
			userIdentifier: identifier,
			permissions:    permissions,
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

func (m *Member) Permissions() []types.Permission {
	return m.permissions
}

func (m *Member) SetPermissions(permissions []types.Permission) {
	m.permissions = permissions
}

func (m *Member) HasPermissions(permissions []types.Permission) bool {
	hasAllRights := true
	for _, wanted := range permissions {
		if !ContainsPermissions(wanted, m.permissions) {
			hasAllRights = false
		}
	}
	return hasAllRights
}

func (m *Member) AddPermissions(permissions []types.Permission) error {
	for _, toAdd := range permissions {
		if !ContainsPermissions(toAdd, m.permissions) {
			m.permissions = append(m.permissions, toAdd)
		}
	}
	return nil
}

func (m *Member) RemovePermissions(permissions []types.Permission) {
	for _, toRemove := range permissions {
		for i, right := range m.permissions {
			if toRemove == right {
				m.permissions = append(m.permissions[:i], m.permissions[i+1:]...)
				break
			}
		}
	}
}

func ContainsPermissions(right types.Permission, permissions []types.Permission) bool {
	contains := false
	for _, r := range permissions {
		if r == right {
			contains = true
		}
	}
	return contains
}

func ValidPermissions(permissions []types.Permission) bool {
	for _, r := range permissions {
		if !ContainsPermissions(r, types.ValidPermissions) {
			return false
		}
	}
	return true
}
