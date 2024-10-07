package propl

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func GetID(conditions MsgCondition, traits SubjectTrait, addons ...string) string {
	s := fmt.Sprintf("%s.%s", conditions.String(), traits.Type().String())
	if len(addons) > 0 {
		for _, a := range addons {
			s = fmt.Sprintf("%s.%s", s, a)
		}
	}
	return s
}

// FaultMap is the result of a policy
// execution
type FaultMap map[string]error

// policy manager maintains the list of configured policies
// and the context with which it was constructed (if supplied).
//
// it also contains a reference to the field store where it can
// fetch policy subjects for evaluation and set them on creation.
type PolicyManager struct {
	ctx            context.Context
	store          PolicySubjectStore
	policies       []TraitPolicy
	pptr           map[string]int
	actionPolicies []ActionPolicy
	aptr           map[string]int
}

// A TraitPolicy is a set of rules that the specified subjects
// must uphold
type TraitPolicy interface {
	ID() string
	EvaluateSubjects(ctx context.Context, subjects ...PolicySubject) FaultMap
}

// An action policy is a custom policy that injects the entire message for
// custom evaluation
type ActionPolicy interface {
	ID() string
	RunAction(ctx context.Context, subject PolicySubject, msg proto.Message) FaultMap
}

func NewPolicyManager(ctx context.Context, store PolicySubjectStore) *PolicyManager {
	return &PolicyManager{
		ctx:   ctx,
		store: store,
		pptr:  make(map[string]int),
		aptr:  make(map[string]int),
	}
}

func (p *PolicyManager) GetTraitPolicy(conditions MsgCondition, traits SubjectTrait) (TraitPolicy, bool) {
	existingIdx, ok := p.pptr[GetID(conditions, traits)]
	if !ok {
		return nil, false
	}
	return p.policies[existingIdx], true
}

func (p *PolicyManager) SetPolicy(conditions MsgCondition, traits SubjectTrait) TraitPolicy {
	policyID := GetID(conditions, traits)
	existingIdx, ok := p.pptr[policyID]
	if ok {
		return p.policies[existingIdx]
	}
	p.policies = append(p.policies, &policy{
		id:         policyID,
		conditions: conditions,
		traits:     traits,
	})
	p.pptr[policyID] = len(p.policies) - 1
	return p.policies[len(p.policies)-1]
}

func (p *PolicyManager) GetActionPolicy(conditions MsgCondition, subject string) (ActionPolicy, bool) {
	existingIdx, ok := p.aptr[GetID(conditions, Traitless(), subject)]
	if !ok {
		return nil, false
	}
	return p.actionPolicies[existingIdx], true
}

func (p *PolicyManager) SetActionPolicy(conditions MsgCondition, action Action, subject string) ActionPolicy {
	policyID := GetID(conditions, Traitless(), subject)
	existingIdx, ok := p.aptr[policyID]
	if ok {
		return p.actionPolicies[existingIdx]
	}
	p.actionPolicies = append(p.actionPolicies, &actionPolicy{
		policy: &policy{
			id:         policyID,
			conditions: conditions,
			traits:     Traitless(),
		},
		a: action,
	})
	p.aptr[policyID] = len(p.actionPolicies) - 1
	return p.actionPolicies[len(p.actionPolicies)-1]
}

// Some function triggered by the result of an evaluation whether it be
// a policy or a global evaluation
type Action func(ctx context.Context, subject PolicySubject, msg proto.Message) error

type PolicyActionOption func(*TraitPolicy)

// a policy subject is a subject that gets evaluated to see:
// 1. what action is configured to occurr if a certain condition is met
// 2. what traits it has if the conditional action results in a trait eval
type PolicySubject interface {
	// some identifier for the subject
	ID() string
	// check for whether or not a policy subject holds a trait
	HasTrait(t SubjectTrait) bool
	// evaluatable reports whether or not the subject is in a state
	// that resolves to some action. for example, if a policy that ensures that
	// a subject has a non-zero value but the evaluation condition is a in the
	// update mask (and the field was not supplied in the mask), the field is not
	// evaluatable
	Evaluatable(conditions MsgCondition, request proto.Message) bool
	// reports whether or not the subject meets the supplied conditions. Anything that
	// returns false for Evaluatable will return false for MeetsConditions
	MeetsConditions(conditions MsgCondition, request proto.Message) bool
}

// a trait is an attribute of a policy subject that must
// be true if the policy is a trait evaluation
type SubjectTrait interface {
	// another trait that must exist with this trait
	And() SubjectTrait
	// another trait that must exist if this trait does not
	Or() SubjectTrait
	// some error string describing the validation error
	UhOhString() string
	// the trait type
	Type() TraitType
	// state check to report the validity of trait
	Valid() bool
}

// a way to store and retrieve policy subjects as well
// as holds a reference to the primary subject
type PolicySubjectStore interface {
	// store the id
	ProcessByID(ctx context.Context, id string) PolicySubject
	// get the policy subject by some id
	GetByID(ctx context.Context, id string) PolicySubject
	// subject is the message that the store was hydrated from
	Source() proto.Message
}
