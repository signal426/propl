package protovalidate

import (
	"google.golang.org/protobuf/proto"
)

func (r *fieldPolicy[T]) neverZero(rpc string, msg proto.Message, paths PathSet) validationFn[T] {
	return func(t T) error {
		isFieldSet := IsFieldSet(msg, r.fullPath)
		if !isFieldSet {
			if r.conditions.Has(Always) && r.conditions.Has(InMask) {
				return mustBeSuppliedInMaskAndNonZero(r.fullPath, rpc)
			}
			if r.conditions.Has(Always) {
				return mustNotBeEmpty(r.fullPath, rpc)
			}
		}
		return nil
	}
}

func (r *requirement[T]) nonZeroIfInMask(rpc string, msg proto.Message, paths PathSet) validationFn[T] {
	return func(t T) error {
		isFieldSet := IsFieldSet(msg, r.fullPath)
		if !isFieldSet {
			if r.conditions.Has(InMask) {
				return mustNotBeEmptyIfInMask(r.fullPath, rpc)
			}
		}
		return nil
	}
}

func (r *requirement[T]) notEqual(rpc string, msg proto.Message, paths PathSet) validationFn[T] {
	return func(t T) error {
		return nil
	}
}
