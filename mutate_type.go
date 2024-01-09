package faketagencoder

import "reflect"

type Skipper func(reflect.Type) bool

func SkipImplementor(rt reflect.Type) Skipper {
	return func(t reflect.Type) bool {
		return t.Implements(rt) ||
			(t.Kind() == reflect.Pointer && t.Elem().Implements(rt)) ||
			reflect.PointerTo(t).Implements(rt)
	}
}

func SkipNot(s Skipper) Skipper {
	return func(t reflect.Type) bool {
		return !s(t)
	}
}

func SkipAnonymous() Skipper {
	return func(t reflect.Type) bool {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Anonymous {
				return true
			}
		}
		return false
	}
}

func CombineSkipper(skippers ...Skipper) Skipper {
	return func(t reflect.Type) bool {
		for _, skipper := range skippers {
			if skipper(t) {
				return true
			}
		}
		return false
	}
}

type TagMutator func(reflect.StructField) reflect.StructTag

func AddOption(tag string, opt string, ignoreIf func(t reflect.Type) bool) TagMutator {
	return func(sf reflect.StructField) reflect.StructTag {
		if ignoreIf(sf.Type) {
			return sf.Tag
		}
		added, err := AddTagOption(sf.Tag, tag, opt)
		if err != nil {
			// downstream may return same error
			return sf.Tag
		}
		return added
	}
}

func MutateTag(
	rt reflect.Type,
	skipAdvancing Skipper,
	mutateTag TagMutator,
) reflect.Type {
	fields := make([]reflect.StructField, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		typ := field.Type

		if !skipAdvancing(typ) {
			if typ.Kind() == reflect.Struct {
				typ = MutateTag(typ, skipAdvancing, mutateTag)
			} else if typ.Kind() == reflect.Pointer {
				elem := typ.Elem()
				if elem.Kind() == reflect.Struct {
					elem = MutateTag(elem, skipAdvancing, mutateTag)
					typ = reflect.PointerTo(elem)
				}
			}
		}

		fields[i] = reflect.StructField{
			Name:      field.Name,
			PkgPath:   field.PkgPath,
			Type:      typ,
			Tag:       mutateTag(field),
			Offset:    field.Offset,
			Index:     field.Index,
			Anonymous: field.Anonymous,
		}
	}

	return reflect.StructOf(fields)
}
