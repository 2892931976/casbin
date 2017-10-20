package rbac

import (
	"github.com/casbin/casbin/util"
)

type sessionRoleManager struct {
	allRoles          map[string]*SessionRole
	maxHierarchyLevel int
}

func SessionRoleManager() RoleManagerConstructor {
	return func() RoleManager {
		return NewSessionRoleManager(10)
	}
}

func NewSessionRoleManager(maxHierarchyLevel int) RoleManager {
	rm := sessionRoleManager{}
	rm.allRoles = make(map[string]*SessionRole)
	rm.maxHierarchyLevel = maxHierarchyLevel
	return &rm
}

func (rm *sessionRoleManager) hasRole(name string) bool {
	_, ok := rm.allRoles[name]
	return ok
}

func (rm *sessionRoleManager) createRole(name string) *SessionRole {
	if !rm.hasRole(name) {
		rm.allRoles[name] = newSessionRole(name)
	}
	return rm.allRoles[name]
}

func (rm *sessionRoleManager) AddLink(name1 string, name2 string, domain ...string) {
	if len(domain) != 2 {
		return
	}
	startTime := domain[0]
	endTime := domain[1]

	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)

	session := Session{role2, startTime, endTime}
	role1.addSession(session)
}

func (rm *sessionRoleManager) DeleteLink(name1 string, name2 string, domain ...string) {
	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return
	}

	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)

	role1.deleteSessions(role2.name)
}

func (rm *sessionRoleManager) HasLink(name1 string, name2 string, requestTime ...string) bool {
	if len(requestTime) != 1 {
		return false
	}

	if name1 == name2 {
		return true
	}

	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return false
	}

	role1 := rm.createRole(name1)
	return role1.hasValidSession(name2, rm.maxHierarchyLevel, requestTime[0])
}

func (rm *sessionRoleManager) GetRoles(name string, domain ...string) []string {
	if !rm.hasRole(name) {
		return nil
	}

	sessionRoles := rm.createRole(name).getSessionRoles()
	return sessionRoles
}

func (rm *sessionRoleManager) GetUsers(name string) []string {
	users := []string{}
	for _, role := range rm.allRoles {
		if role.hasDirectRole(name) {
			users = append(users, role.name)
		}
	}
	return users
}

func (rm *sessionRoleManager) PrintRoles() {
	for _, role := range rm.allRoles {
		util.LogPrint(role.toString())
	}
}

type SessionRole struct {
	name     string
	sessions []Session
}

func newSessionRole(name string) *SessionRole {
	sr := SessionRole{name: name}
	return &sr
}

func (sr *SessionRole) addSession(s Session) {
	sr.sessions = append(sr.sessions, s)
}

func (sr *SessionRole) deleteSessions(sessionName string) {
	// Delete sessions from an array while iterating it
	index := 0
	for _, srs := range sr.sessions {
		if srs.role.name == sessionName {
			sr.sessions[index] = srs
			index++
		}
	}
	sr.sessions = sr.sessions[:index]
}

func (sr *SessionRole) getSessions() []Session {
	return sr.sessions
}

func (sr *SessionRole) getSessionRoles() []string {
	names := []string{}
	for _, session := range sr.sessions {
		names = append(names, session.role.name)
	}
	return names
}

func (sr *SessionRole) hasValidSession(name string, hierarchyLevel int, requestTime string) bool {
	if hierarchyLevel == 0 {
		return false
	}

	for _, s := range sr.sessions {
		if s.startTime <= requestTime && requestTime <= s.endTime {
			if s.role.name == name {
				return true
			}
			if s.role.hasValidSession(name, hierarchyLevel-1, requestTime) {
				return true
			}
		}
	}
	return false
}

func (sr *SessionRole) hasDirectRole(name string) bool {
	for _, session := range sr.sessions {
		if session.role.name == name {
			return true
		}
	}
	return false
}

func (sr *SessionRole) toString() string {
	sessions := ""
	for i, session := range sr.sessions {
		if i == 0 {
			sessions += session.role.name
		} else {
			sessions += ", " + session.role.name
		}
		sessions += " (until: " + session.endTime + ")"
	}
	return sr.name + " < " + sessions
}

type Session struct {
	role      *SessionRole
	startTime string
	endTime   string
}
