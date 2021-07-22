package jamsession

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
)

type Members map[string]Member

type Member struct {
	User *types.User
	Rights []types.MemberRights
}

func (m Members) Get (user *types.User) (*Member, error) {
	if member, ok := m[user.Identifier]; ok {
		return &member, nil
	}
	return nil, errors.New("not a member")
}

func (m Members) Add (user *types.User, rights []types.MemberRights) bool {
	if _, ok := m[user.Identifier]; !ok {
		m[user.Identifier] = Member{
			User: user,
			Rights: rights,
		}
		return true
	}
	return false
}

func (m Members) Remove (user *types.User) bool {
	if _, ok := m[user.Identifier]; ok {
		delete(m, user.Identifier)
		return true
	}
	return false
}

func (m *Member) Has(rights []types.MemberRights) bool {
	hasAll := true
	for _, wanted := range rights {
		has := false
		for _, right := range m.Rights {
			if right == wanted {
				has = true
			}
		}
		if !has {
			hasAll = false
		}
	}
	return hasAll
}

func (m *Member) Add(rights []types.MemberRights) {
	for _, toAdd := range rights {
		if !m.Has([]types.MemberRights{toAdd}) {
			m.Rights = append(m.Rights, toAdd)
		}
	}
}

func (m *Member) Remove(rights []types.MemberRights) bool {
	for _, toRemove := range rights {
		for i, right := range m.Rights {
			if toRemove == right {
				m.Rights = append(m.Rights[:i], m.Rights[i+1:]...)
				return true
			}
		}
	}
	return false
}