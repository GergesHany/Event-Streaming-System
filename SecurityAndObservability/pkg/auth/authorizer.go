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
   enforcer := casbin.NewEnforcer(model, policy)
   return &Authorizer{enforcer: enforcer}
}

func (a *Authorizer) Authorize(subject, object, action string) (error) {
	// Check permission using Casbin enforcer
	if !a.enforcer.Enforce(subject, object, action) {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}