package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

const (
	prevalErr  = "the message failed prevalidation"
	postValErr = "the message failed postvalidation"
)

// output format of the errors
type Format uint32

const (
	Default Format = iota
	JSON
)

// Represents a field id and a condition that triggers
// an evaluation of said field
type FieldCondition struct {
	Field     string
	Value     interface{}
	Condition MsgCondition
}

// IsInMask constructs a condition that dictates the field
// is only evaluated if it is speficied in an update mask
func IsInMask(field string, value interface{}) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask,
		Value:     value,
	}
}

// Always constructs a condition that dictates the field
// is always expected to be present for evaluation
func AlwaysPresent(field string, value interface{}) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask.And(InMessage),
		Value:     value,
	}
}

type SubjectUnderEvaluation struct {
	// custom evaluations triggered by a field condition.
	// custom evaluation callbacks are provided the entire payload, not just the field data.
	// these callbacks get triggererd before any policies have been evaluated.
	// they are executed in the order in which they were attached to the subject.
	prepoliciesactions []Action
	// custom evaluations triggered by a field condition. these
	// callbacks get triggered after all policies have been evaluated.
	postpoliciesactions []Action
	// policies is the map of field ids to some policy configuration
	pm *PolicyManager
	// map of trait policy ids to a list of subject ids
	traitPolicySubjectMap map[string][]string
	// map of action policy ids to a list of subject ids
	actionPolicySubjectMap map[string][]string
	// pointer to policy index for faster lookups
	// a generator of policies traditional field policies
	// pointer to policy index for faster lookups
	apdx map[string]int
	// errResultHandler is the handler that manages the output of the evaluations
	uhoh UhOhHandler
	// paths is list of fields that are being evaluated if a field mask is supplied
	paths []string
	// indicator if the evaluator should exit if any of the pre-functions fail.
	// the default behavior is that this is false
	continueOnGlobalEvalErr bool
}

// Options to provide to the subject evaluation
type EvaluationOption func(*SubjectUnderEvaluation)

// Optionally provide a ctx object. This will be passed on
// any global callback.
func WithCtx(ctx context.Context) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		s.ctx = ctx
	}
}

// Indicator to keep going if a global check fails. Otherwuse
// will exit the remaining validation after the error. Off by default.
func WithContinueOnGlobalEvalErr() EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		s.continueOnGlobalEvalErr = true
	}
}

// Specify global pre-checks. These are executed in the order in which they
// are specified.
func WithPrePolicyAction(e Action) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		if s.prepoliciesactions == nil {
			s.prepoliciesactions = make([]Action, 0, 3)
		}
		s.prepoliciesactions = append(s.prepoliciesactions, e)
	}
}

// Specify global pos-checks. These are executed inthe order in which they
// are specified
func WithPostPolicyAction(e Action) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		if s.postpoliciesactions == nil {
			s.postpoliciesactions = make([]Action, 0, 3)
		}
		s.postpoliciesactions = append(s.postpoliciesactions, e)
	}
}

// Specifies update paths for eval subject. This drives the conditional
// assertions
func WithMaskPaths(paths ...string) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		s.paths = paths
	}
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func ForSubject(subject proto.Message, options ...EvaluationOption) *SubjectUnderEvaluation {
	s := &SubjectUnderEvaluation{}
	if len(options) > 0 {
		for _, o := range options {
			o(s)
		}
	}
	return s
}

// HasNonZeroField pass in a list of fields that must not be equal to their
// zero value
//
// example: sue := HasNonZeroFields("user.id", "user.first_name")
func (p *SubjectUnderEvaluation) AssertNonZero(path string, value interface{}) *SubjectUnderEvaluation {
	var (
		policy TraitPolicy
		ok     bool
	)
	policy, ok = p.pm.GetTraitPolicy(AlwaysInMsg(), NotZeroTrait())
	if !ok {
		p.pm.SetPolicy(AlwaysInMsg(), NotZeroTrait())
	}
	return p
}

// HasNonZeroFieldsWhen pass in a list of field conditions if you want to customize the conditions under which
// a field non-zero evaluation is triggered
//
// example: sue := HasNonZeroFieldsWhen(IfInMask("user.first_name"), Always("user.first_name"))
func (p *SubjectUnderEvaluation) HasNonZeroFieldsWhen(conds ...FieldCondition) *SubjectUnderEvaluation {
	for _, c := range conds {
		id := GetID(c.Condition, NotZeroTrait())
		policy, ok := p.pm.GetTraitPolicy(c.Condition, NotZeroTrait())
		if !ok {
			policy = p.pm.SetPolicy(c.Condition, NotZeroTrait())
			p.traitPolicySubjectMap[id] = []string{c.Field}
		} else {
			p.traitPolicySubjectMap[policy.ID()] = append(p.traitPolicySubjectMap[policy.ID()], c.Field)
		}
	}
	return p
}

// HasCustomEvaluation sets the specified evaluation on the field and will be run if the conditions are met.
func (p *SubjectUnderEvaluation) HasCustomEvaluation(field string, action Action) *SubjectUnderEvaluation {
	policy := p.pm.SetActionPolicy(AlwaysInMsg(), action, field)
	_, ok := p.pm.GetActionPolicy(AlwaysInMsg(), field)
	p.actionPolicySubjectMap[policy.ID()] = append(p.actionPolicySubjectMap[policy.ID()], field)
	if !ok {
		p.apdx[policy.ID()] = len(p.actionPolicySubjectMap) - 1
	}
	return p
}

// HasCustomEvaluationWhen sets the specified evaluation on the field and will be run if the conditions are met
func (p *SubjectUnderEvaluation) HasCustomEvaluationWhen(conditions FieldCondition, eval Action) *SubjectUnderEvaluation {
	policySubject := p.store.ProcessByID(p.ctx, conditions.Field)
	if existingidx, ok := p.apdx[conditions.Field]; ok {
		existing := p.actionPolicies[existingidx]
		existing.Subjects = []PolicySubject{policySubject}
		existing.Traits = &trait{traitType: Custom}
		existing.Conditions = conditions.Condition
		existing.Action = eval
	} else {
		newPolicy := &ActionPolicy{
			Policy: &TraitPolicy{
				Subjects:   []PolicySubject{policySubject},
				Traits:     &trait{traitType: Custom},
				Conditions: conditions.Condition,
			},
			Action: eval,
		}
		p.actionPolicies = append(p.actionPolicies, newPolicy)
		p.apdx[conditions.Field] = len(p.actionPolicies) - 1
	}
	return p
}

// CustomErrResultHandler call this before calling E() or Evaluate() if you want to override
// the errors that are output from the policy execution
func (s *SubjectUnderEvaluation) CustomUhOhHandler(e UhOhHandler) *SubjectUnderEvaluation {
	s.uhoh = e
	return s
}

// E shorthand for Evaluate
func (s *SubjectUnderEvaluation) E() error {
	return s.Evaluate()
}

func (s *SubjectUnderEvaluation) handlePrevals() ([]UhOh, error) {
	var uhohs []UhOh
	// evaluate the global pre-checks
	if s.prepoliciesactions != nil && len(s.prepoliciesactions) > 0 {
		for _, action := range s.prepoliciesactions {
			err := action(s.ctx, s.store.Subject())
			if err != nil {
				if !s.continueOnGlobalEvalErr {
					return nil, s.uhoh.SpaghettiOs(uhohs)
				}
				uhohs = append(uhohs, RequestUhOh(err))
			}
		}
	}
	return uhohs, nil
}

func (s *SubjectUnderEvaluation) handlePostVals() ([]UhOh, error) {
	// instantiate because empty actually has meaning
	uhohs := []UhOh{}
	// evaluate the global pre-checks
	if s.postpoliciesactions != nil && len(s.postpoliciesactions) > 0 {
		for _, action := range s.postpoliciesactions {
			err := action(s.ctx, s.store.Subject())
			if err != nil {
				if !s.continueOnGlobalEvalErr {
					return nil, s.uhoh.SpaghettiOs(uhohs)
				}
				uhohs = append(uhohs, RequestUhOh(err))
			}
		}
	}
	return uhohs, nil
}

// Evaluate checks each declared policy and returns an error describing
// each infraction. If a precheck is specified and returns an error, this exits
// and field policies are not evaluated.
//
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (s *SubjectUnderEvaluation) Evaluate() error {
	uhohs, err := s.handlePrevals()
	if err != nil {
		return err
	}
	for _, policy := range s.policies {
		if errs := policy.EvaluateSubject(s.ctx, s.store.Subject()); err != nil {
			for field, err := range errs {
				uhohs = append(uhohs, FieldUhOh(field, err))
			}
		}
	}
	postuhohs, err := s.handlePostVals()
	if err != nil {
		return err
	}
	uhohs = append(uhohs, postuhohs...)
	return s.uhoh.SpaghettiOs(uhohs)
}
