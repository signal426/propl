package propl

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// our traditional policy with sets of conditions and traits that
// must be upheld
var _ Policy = (*propolicy)(nil)

type propolicy struct {
	conditions Condition
	traits     Trait
}

// Execute checks traits on the field based on the conditional action signal
// returned from the subject.
func (p *propolicy) EvaluateSubject(ctx context.Context, subject PolicySubject) error {
	if !subject.Evaluatable(p.conditions) {
		return nil
	}
	if !subject.MeetsConditions(p.conditions) {
		return fmt.Errorf("subject did not meet conditions %s", p.conditions.FlagsString())
	}
	return p.evaluateSubjectTraits(subject)
}

func (p *propolicy) evaluateSubjectTraits(subject PolicySubject) error {
	return p.checkTraits(subject, p.traits)
}

func (p *propolicy) checkTraits(s PolicySubject, t Trait) error {
	if t == nil {
		return nil
	}
	if t.Valid() && !s.HasTrait(t) {
		// if we have an or, keep going
		if t.Or().Valid() {
			return p.checkTraits(s, t.Or())
		}
		// else, we're done checking
		return errors.New(t.UhOhString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if t.And().Valid() {
		return p.checkTraits(s, t.And())
	}
	return nil
}

// custom policy that executes a function that receives the
// the entire aggregate object whenever a subject condition is met
var _ Policy = (*customPropolicy)(nil)

type customPropolicy struct {
	arg        proto.Message
	conditions Condition
	eval       TriggeredEvaluation
}

func (p *customPropolicy) EvaluateSubject(ctx context.Context, subject PolicySubject) error {
	if !subject.Evaluatable(p.conditions) {
		return nil
	}
	if !subject.MeetsConditions(p.conditions) {
		return fmt.Errorf("subject did not meet conditions %s", p.conditions.FlagsString())
	}
	return p.triggeredEvaluation(ctx)
}

func (mp *customPropolicy) triggeredEvaluation(ctx context.Context) error {
	return mp.eval(ctx, mp.arg)
}
