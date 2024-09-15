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

// Represents a field id and a condition that triggers
// an evaluation of said field
type FieldCondition struct {
	Field     string
	Condition Condition
}

// IsInMask constructs a condition that dictates the field
// is only evaluated if it is speficied in an update mask
func IsInMask(field string) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask,
	}
}

// Always constructs a condition that dictates the field
// is always expected to be present for evaluation
func Always(field string) FieldCondition {
	return FieldCondition{
		Field:     field,
		Condition: InMask.And(InMessage),
	}
}

type SubjectUnderEvaluation struct {
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
func WithPrePolicyEvaluation(e TriggeredEvaluation) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		if s.customEvaluationsPrePol == nil {
			s.customEvaluationsPrePol = make([]TriggeredEvaluation, 0, 3)
		}
		s.customEvaluationsPrePol = append(s.customEvaluationsPrePol, e)
	}
}

// Specify global pos-checks. These are executed inthe order in which they
// are specified
func WithPostPolicyEvaluation(e TriggeredEvaluation) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		if s.customEvaluationsPostPol == nil {
			s.customEvaluationsPostPol = make([]TriggeredEvaluation, 0, 3)
		}
		s.customEvaluationsPostPol = append(s.customEvaluationsPostPol, e)
	}
}

// Specifies update paths for eval subject. This drives the conditional
// assertions
func WithMaskPaths(paths ...string) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		s.paths = paths
	}
}

func WithFieldStore(fs PolicySubjectStore) EvaluationOption {
	return func(s *SubjectUnderEvaluation) {
		s.store = fs
	}
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func ForSubject(subject proto.Message, options ...EvaluationOption) *SubjectUnderEvaluation {
	s := &SubjectUnderEvaluation{
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

// HasNonZeroFields pass in a list of fields that must not be equal to their
// zero value
//
// example: sue := HasNonZeroFields("user.id", "user.first_name")
func (p *SubjectUnderEvaluation) HasNonZeroFields(fields ...string) *SubjectUnderEvaluation {
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

// HasNonZeroFieldsWhen pass in a list of field conditions if you want to customize the conditions under which
// a field non-zero evaluation is triggered
//
// example: sue := HasNonZeroFieldsWhen(IfInMask("user.first_name"), Always("user.first_name"))
func (p *SubjectUnderEvaluation) HasNonZeroFieldsWhen(conds ...FieldCondition) *SubjectUnderEvaluation {
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

// HasCustomEvaluation sets the specified evaluation on the field and will be run if the conditions are met.
func (p *SubjectUnderEvaluation) HasCustomEvaluation(field string, eval TriggeredEvaluation) *SubjectUnderEvaluation {
	policy := &customPropolicy{
		conditions: InMask.And(InMessage),
		arg:        p.store.Subject(),
		eval:       eval,
	}
	p.policies[policy] = append(p.policies[policy], p.store.ProcessByID(p.ctx, field))
	return p
}

// HasCustomEvaluationWhen sets the specified evaluation on the field and will be run if the conditions are met
func (p *SubjectUnderEvaluation) HasCustomEvaluationWhen(conditions FieldCondition, eval TriggeredEvaluation) *SubjectUnderEvaluation {
	policy := &customPropolicy{
		conditions: conditions.Condition,
		arg:        p.store.Subject(),
		eval:       eval,
	}
	f := p.store.ProcessByID(p.ctx, conditions.Field)
	p.policies[policy] = append(p.policies[policy], f)
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

// Evaluate checks each declared policy and returns an error describing
// each infraction. If a precheck is specified and returns an error, this exits
// and field policies are not evaluated.
//
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (s *SubjectUnderEvaluation) Evaluate() error {
	uhohs := []UhOh{}
	// evaluate the global pre-checks
	if s.customEvaluationsPrePol != nil && len(s.customEvaluationsPrePol) > 0 {
		for _, pvc := range s.customEvaluationsPrePol {
			err := pvc(s.ctx, s.store.Subject())
			if err != nil {
				if s.continueOnGlobalEvalErr {
					uhohs = append(uhohs, RequestUhOh(err))
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
				uhohs = append(uhohs, FieldUhOh(subject.ID(), err))
			}
		}
	}
	// evaluate the global post-checks
	if s.customEvaluationsPostPol != nil && len(s.customEvaluationsPostPol) > 0 {
		for _, pvc := range s.customEvaluationsPostPol {
			err := pvc(s.ctx, s.store.Subject())
			if err != nil {
				if s.continueOnGlobalEvalErr {
					uhohs = append(uhohs, RequestUhOh(err))
				} else {
					return err
				}
			}
		}
	}
	if len(uhohs) > 0 {
		if s.uhoh == nil {
			s.uhoh = newDefaultUhOhHandler()
		}
		return s.uhoh.SpaghettiOs(uhohs)
	}
	return nil
}
