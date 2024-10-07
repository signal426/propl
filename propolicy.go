package propl

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ TraitPolicy = (*policy)(nil)

type policy struct {
	id         string
	traits     SubjectTrait
	conditions MsgCondition
	store      PolicySubjectStore
}

func (p policy) ID() string {
	return p.id
}

// EvaluateSubjects implements Policy.
func (p *policy) EvaluateSubjects(ctx context.Context, subjects ...PolicySubject) FaultMap {
	errs := map[string]error{}
	for _, subject := range subjects {
		if !subject.Evaluatable(p.conditions, p.store.Source()) {
			return nil
		}
		if !subject.MeetsConditions(p.conditions, p.store.Source()) {
			errs[subject.ID()] = fmt.Errorf("%s did not meet conditions %s", subject.ID(), p.conditions.FlagsString())
		} else if err := p.assertSubjectTraits(subject, p.traits); err != nil {
			errs[subject.ID()] = err
		}
	}
	return errs
}

func (p *policy) assertSubjectTraits(subject PolicySubject, mustHave SubjectTrait) error {
	if mustHave == nil {
		return nil
	}
	if mustHave.Valid() && !subject.HasTrait(mustHave) {
		// if we have an or, keep going
		if mustHave.Or().Valid() {
			return p.assertSubjectTraits(subject, mustHave.Or())
		}
		// else, we're done checking
		return errors.New(mustHave.UhOhString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if mustHave.And().Valid() {
		return p.assertSubjectTraits(subject, mustHave.And())
	}
	return nil
}

var _ ActionPolicy = (*actionPolicy)(nil)

type actionPolicy struct {
	*policy
	a Action
}

func (p actionPolicy) ID() string {
	return p.id
}

func (p *actionPolicy) RunAction(ctx context.Context, subject PolicySubject, msg protoreflect.ProtoMessage) FaultMap {
	errs := map[string]error{}
	if !subject.Evaluatable(p.conditions, p.store.Source()) {
		return nil
	}
	if !subject.MeetsConditions(p.conditions, p.store.Source()) {
		errs[subject.ID()] = fmt.Errorf("%s did not meet conditions %s", subject.ID(), p.conditions.FlagsString())
	} else if err := p.a(ctx, subject, p.store.Source()); err != nil {
		errs[subject.ID()] = err
	}
	return errs
}
