package main

import (
	"fmt"
	"reflect"

	"github.com/go-json-experiment/json" // github.com/go-json-experiment/json v0.0.0-20231102232822-2e55bd4e08b0
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/faketagencoder"
)

type opt[V any] struct {
	valid bool
	v     V
}

type und[V any] struct {
	opt opt[opt[V]]
}

func Undefined[V any]() und[V] {
	return und[V]{}
}

func Null[V any]() und[V] {
	return und[V]{
		opt: opt[opt[V]]{
			valid: true,
		},
	}
}

func Defined[V any](v V) und[V] {
	return und[V]{
		opt: opt[opt[V]]{
			valid: true,
			v: opt[V]{
				valid: true,
				v:     v,
			},
		},
	}
}

func (u *und[V]) IsZero() bool {
	return u.IsUndefined()
}

func (u *und[V]) IsUndefined() bool {
	return !u.opt.valid
}

func (u *und[V]) IsNull() bool {
	return !u.IsUndefined() && !u.opt.v.valid
}

func (u *und[V]) Value() V {
	if u.IsUndefined() || u.IsNull() {
		var zero V
		return zero
	}
	return u.opt.v.v
}

var _ json.MarshalerV2 = (*und[any])(nil)

func (u *und[V]) MarshalJSONV2(enc *jsontext.Encoder, opt json.Options) error {
	if u.IsUndefined() || u.IsNull() {
		return enc.WriteToken(jsontext.Null)
	}
	return json.MarshalEncode(enc, u.Value(), opt)
}

var _ json.UnmarshalerV2 = (*und[any])(nil)

func (u *und[V]) UnmarshalJSONV2(dec *jsontext.Decoder, opt json.Options) error {
	var v V
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		u.opt.valid = true
		u.opt.v.valid = false
		u.opt.v.v = v
		return nil
	}
	err := json.UnmarshalDecode(dec, &v)
	if err != nil {
		return err
	}
	u.opt.valid = true
	u.opt.v.valid = true
	u.opt.v.v = v
	return nil
}

type Undefinedable interface {
	IsUndefined() bool
}

var (
	undefinedableType = reflect.TypeOf((*Undefinedable)(nil)).Elem()
	jsonV1Marshaller  = reflect.TypeOf((*json.MarshalerV1)(nil)).Elem()
	jsonV2Marshaller  = reflect.TypeOf((*json.MarshalerV2)(nil)).Elem()
)

func main() {
	type Nested struct {
		Nah und[string] `json:"nah"`
		Yay int         `json:"yay"`
	}
	type some struct {
		Foo und[string] `json:"foo"`
		Bar string      `json:"bar"`
		Baz Nested      `json:"baz"`
		Nested
	}
	mutaed := faketagencoder.MutateTag(
		reflect.TypeOf(some{}),
		faketagencoder.CombineSkipper(
			faketagencoder.SkipImplementor(jsonV1Marshaller),
			faketagencoder.SkipImplementor(jsonV2Marshaller),
		),
		faketagencoder.AddOption(`json`, `,omitzero`, faketagencoder.SkipNot(faketagencoder.SkipImplementor(undefinedableType))),
	)

	fmt.Printf("mutaed type = %+v\n", mutaed)

	marshaller := json.MarshalFuncV2[some](func(e *jsontext.Encoder, s some, o json.Options) error {
		rv := reflect.ValueOf(s)
		v := reflect.New(mutaed).Elem()
		setExported(v, rv)
		return json.MarshalEncode(e, v.Interface(), o)
	})

	for _, v := range []some{
		{},
		{
			Foo: Defined("foo"),
			Bar: "bar",
			Baz: Nested{
				Nah: Defined("nah"),
				Yay: 20,
			},
			Nested: Nested{
				Nah: Defined("nah"),
				Yay: -231,
			},
		},
	} {
		out, err := json.Marshal(v, json.WithMarshalers(marshaller), jsontext.WithIndent("    "))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", out)
	}
	/*
		{
		    "bar": "",
		    "baz": {
		        "yay": 0
		    },
		    "yay": 0
		}
		{
		    "foo": "foo",
		    "bar": "bar",
		    "baz": {
		        "nah": "nah",
		        "yay": 20
		    },
		    "nah": "nah",
		    "yay": -231
		}
	*/
}

func setExported(l, r reflect.Value) {
	for i := 0; i < l.NumField(); i++ {
		fl := l.Field(i)
		fr := r.Field(i)
		if fl.Type() == fr.Type() {
			fl.Set(fr)
		} else {
			setExported(fl, fr)
		}
	}
}

/*
bin = {"Foo":"foo","Bar":"bar"}, err = <nil>
value = foo, undefined = false, null = false, err = <nil>
bin = {"Foo":"","Bar":"bar"}, err = <nil>
value = , undefined = false, null = false, err = <nil>
bin = {"Foo":null,"Bar":"bar"}, err = <nil>
value = , undefined = false, null = true, err = <nil>
bin = {"Bar":"bar"}, err = <nil>
value = , undefined = true, null = false, err = <nil>
*/
