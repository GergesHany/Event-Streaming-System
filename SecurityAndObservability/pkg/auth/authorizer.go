package auth

import (
	"fmt"

	"github.com/casbin/casbin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	// Casbin enforcer for authorization
	enforcer *casbin.Enforcer
}

func New(model, policy string) *Authorizer {
	// If model or policy is empty, create a permissive enforcer (allow all)
	if model == "" || policy == "" {
		return &Authorizer{enforcer: nil}
	}
	enforcer := casbin.NewEnforcer(model, policy)
	return &Authorizer{enforcer: enforcer}
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	// If no enforcer is configured, allow all requests
	if a.enforcer == nil {
		return nil
	}
	// Check permission using Casbin enforcer
	if !a.enforcer.Enforce(subject, object, action) {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}
