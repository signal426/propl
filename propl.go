package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

const (
	prevalErr  = "the message failed prevalidation"
	postValErr = "the message failed postvalidation"
	globalKey  = "global"
)

type FieldCondition struct {
	Field     string
	Condition Condition
}

func IsInMask(field string) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask,
	}
}

func Always(field string) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask.And(InMessage),
	}
}

type EvaluationSubject struct {
	// context if supplied
	ctx context.Context
	// the store for the policy subjects that compose this evaluation
	store PolicySubjectStore
	// custom evaluations triggered by a field condition.
	// custom evaluation callbacks are provided the entire payload, not just the field data.
	// these callbacks get triggererd before any policies have been evaluated.
	// they are executed in the order in which they were attached to the subject.
	customEvaluationsPrePol []TriggeredEvaluation
	// custom evaluations triggered by a field condition. these
	// callbacks get triggered after all policies have been evaluated.
	customEvaluationsPostPol []TriggeredEvaluation
	// policies is the map of field ids to some policy configuration
	policies map[Policy][]PolicySubject
	// errResultHandler is the handler that manages the output of the evaluations
	errResultHandler ErrResultHandler
	// paths is list of fields that are being evaluated if a field mask is supplied
	paths []string
	// indicator if the evaluator should exit if any of the pre-functions fail.
	// the default behavior is that this is false
	continueOnGlobalEvalErr bool
}

// Options to provide to the subject evaluation
type EvaluationOption func(*EvaluationSubject)

// Optionally provide a ctx object. This will be passed on
// any global callback.
func WithCtx(ctx context.Context) EvaluationOption {
	return func(s *EvaluationSubject) {
		s.ctx = ctx
	}
}

// Indicator to keep going if a global check fails. Otherwuse
// will exit the remaining validation after the error. Off by default.
func WithContinueOnGlobalEvalErr() EvaluationOption {
	return func(s *EvaluationSubject) {
		s.continueOnGlobalEvalErr = true
	}
}

// Specify global pre-checks. These are executed in the order in which they
// are specified.
func WithPrePolicyEvaluation(e TriggeredEvaluation) EvaluationOption {
	return func(s *EvaluationSubject) {
		if s.customEvaluationsPrePol == nil {
			s.customEvaluationsPrePol = make([]TriggeredEvaluation, 0, 3)
		}
		s.customEvaluationsPrePol = append(s.customEvaluationsPrePol, e)
	}
}

// Specify global pos-checks. These are executed inthe order in which they
// are specified
func WithPostPolicyEvaluation(e TriggeredEvaluation) EvaluationOption {
	return func(s *EvaluationSubject) {
		if s.customEvaluationsPostPol == nil {
			s.customEvaluationsPostPol = make([]TriggeredEvaluation, 0, 3)
		}
		s.customEvaluationsPostPol = append(s.customEvaluationsPostPol, e)
	}
}

// Specifies update paths for eval subject. This drives the conditional
// assertions
func WithMaskPaths(paths ...string) EvaluationOption {
	return func(s *EvaluationSubject) {
		s.paths = paths
	}
}

func WithFieldStore(fs PolicySubjectStore) EvaluationOption {
	return func(s *EvaluationSubject) {
		s.store = fs
	}
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func ForSubject(subject proto.Message, options ...EvaluationOption) *EvaluationSubject {
	s := &EvaluationSubject{
		policies: make(map[Policy][]PolicySubject),
	}
	if len(options) > 0 {
		for _, o := range options {
			o(s)
		}
	}
	if s.store == nil {
		s.store = initFieldStore(subject, withPaths(s.paths...))
	}
	return s
}

func (p *EvaluationSubject) HasNonZeroFields(fields ...string) *EvaluationSubject {
	policy := &propolicy{
		conditions: InMask.And(InMessage),
		traits: &trait{
			traitType: NotZero,
		},
	}
	_, ok := p.policies[policy]
	if !ok {
		p.policies[policy] = make([]PolicySubject, 0, 3)
	}
	for _, f := range fields {
		p.policies[policy] = append(p.policies[policy], p.store.ProcessByID(p.ctx, f))
	}
	return p
}

func (p *EvaluationSubject) HasNonZeroFieldsWhen(conds ...FieldCondition) *EvaluationSubject {
	for _, c := range conds {
		policy := &propolicy{
			conditions: c.Condition,
			traits: &trait{
				traitType: NotZero,
			},
		}
		_, ok := p.policies[policy]
		if !ok {
			p.policies[policy] = make([]PolicySubject, 0, 3)
		}
		p.policies[policy] = append(p.policies[policy], p.store.ProcessByID(p.ctx, c.Field))
	}
	return p
}

func (p *EvaluationSubject) HasCustomEvaluation(field string, eval TriggeredEvaluation) *EvaluationSubject {
	policy := &customPropolicy{
		conditions: InMask.And(InMessage),
		arg:        p.store.Subject(),
		eval:       eval,
	}
	p.policies[policy] = append(p.policies[policy], p.store.ProcessByID(p.ctx, field))
	return p
}

func (p *EvaluationSubject) HasCustomEvaluationWhen(conditions FieldCondition, eval TriggeredEvaluation) *EvaluationSubject {
	policy := &customPropolicy{
		conditions: conditions.Condition,
		arg:        p.store.Subject(),
		eval:       eval,
	}
	f := p.store.ProcessByID(p.ctx, conditions.Field)
	p.policies[policy] = append(p.policies[policy], f)
	return p
}

func (s *EvaluationSubject) CustomErrResultHandler(e ErrResultHandler) *EvaluationSubject {
	s.errResultHandler = e
	return s
}

// E shorthand for Evaluate
func (s *EvaluationSubject) E() error {
	return s.Evaluate()
}

// Evaluate checks each declared policy and returns an error describing
// each infraction. If a precheck is specified and returns an error, this exits
// and field policies are not evaluated.
//
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (s *EvaluationSubject) Evaluate() error {
	finfractions := make(map[string]error)
	// evaluate the global pre-checks
	if s.customEvaluationsPrePol != nil && len(s.customEvaluationsPrePol) > 0 {
		for _, pvc := range s.customEvaluationsPrePol {
			err := pvc(s.ctx, s.store.Subject())
			if err != nil {
				if s.continueOnGlobalEvalErr {
					finfractions[globalKey] = err
				} else {
					return err
				}
			}
		}
	}
	// evaluate the list of policies
	for policy, subjects := range s.policies {
		for _, subject := range subjects {
			if err := policy.EvaluateSubject(s.ctx, subject); err != nil {
				finfractions[subject.ID()] = err
			}
		}
	}
	// evaluate the global post-checks
	if s.customEvaluationsPostPol != nil && len(s.customEvaluationsPostPol) > 0 {
		for _, pvc := range s.customEvaluationsPostPol {
			err := pvc(s.ctx, s.store.Subject())
			if err != nil {
				if s.continueOnGlobalEvalErr {
					finfractions[globalKey] = err
				} else {
					return err
				}
			}
		}
	}
	if len(finfractions) > 0 {
		if s.errResultHandler == nil {
			s.errResultHandler = newDefaultErrResultHandler()
		}
		return s.errResultHandler.Process(finfractions)
	}
	return nil
}
