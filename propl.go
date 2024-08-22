package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type FC struct {
	Field     string
	Condition Condition
}

func IsInMask(field string) FC {
	return FC{
		Field:     field,
		Condition: InMask,
	}
}

func Always(field string) FC {
	return FC{
		Field:     field,
		Condition: InMask.And(InMessage),
	}
}

type EvaluationSubject[T proto.Message] struct {
	store            *fieldStore[T]
	beforeFields     BeforeFields[T]
	policies         map[string]Policy
	errResultHandler ErrResultHandler
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func Subject[T proto.Message](subject T, paths ...string) *EvaluationSubject[T] {
	return &EvaluationSubject[T]{
		store: newFieldStore(subject, paths...),
	}
}

func (p *EvaluationSubject[T]) BeforeFields(b BeforeFields[T]) *EvaluationSubject[T] {
	p.beforeFields = b
	return p
}

func (p *EvaluationSubject[T]) NeverZero(fields ...string) *EvaluationSubject[T] {
	for _, f := range fields {
		fp := &policy{
			subject:    p.store.processPath(f).dataAtPath(f),
			conditions: InMask.And(InMessage),
			traits: &trait{
				traitType: NotZero,
			},
		}
		p.policies[f] = fp
	}
	return p
}

func (p *EvaluationSubject[T]) NeverZeroWhen(conds ...FC) *EvaluationSubject[T] {
	for _, c := range conds {
		fp := &policy{
			subject:    p.store.processPath(c.Field).dataAtPath(c.Field),
			conditions: c.Condition,
			traits: &trait{
				traitType: NotZero,
			},
		}
		p.policies[c.Field] = fp
	}
	return p
}

func (p *EvaluationSubject[T]) NeverErr(field string, fn func(msg T) error) *EvaluationSubject[T] {
	fp := &customPolicy[T]{
		conditions: InMask.And(InMessage),
		arg:        p.store.message(),
		subject:    p.store.processPath(field).dataAtPath(field),
		eval:       fn,
	}
	p.policies[field] = fp
	return p
}

func (p *EvaluationSubject[T]) NeverErrWhen(conditions FC, fn func(msg T) error) *EvaluationSubject[T] {
	fp := &customPolicy[T]{
		conditions: conditions.Condition,
		arg:        p.store.message(),
		subject:    p.store.processPath(conditions.Field).dataAtPath(conditions.Field),
		eval:       fn,
	}
	p.policies[conditions.Field] = fp
	return p
}

func (s *EvaluationSubject[T]) CustomErrResultHandler(e ErrResultHandler) *EvaluationSubject[T] {
	s.errResultHandler = e
	return s
}

// E shorthand for Evaluate
func (s *EvaluationSubject[T]) E(ctx context.Context) error {
	return s.Evaluate(ctx)
}

// Evaluate checks each declared policy and returns an error describing
// each infraction. If a precheck is specified and returns an error, this exits
// and field policies are not evaluated.
//
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (s *EvaluationSubject[T]) Evaluate(ctx context.Context) error {
	if s.beforeFields != nil {
		if err := s.beforeFields(ctx, s.store.message()); err != nil {
			return err
		}
	}
	finfractions := make(map[string]error)
	for id, p := range s.policies {
		if err := p.Execute(); err != nil {
			finfractions[id] = err
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
