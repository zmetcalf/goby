package vm

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/goby-lang/goby/vm/classes"
	"github.com/goby-lang/goby/vm/errors"
)

// HashObject represents hash instances
// Hash is a collection of key-value pair, which works like a dictionary.
// Hash literal is represented with curly brackets `{ }` like `{ key: value }`.
// Each key of the hash is unique and cannot be duplicate within the hash.
// Adding a leading space and a trailing space within curly brackets are preferable.
//
// - **Key:** an alphanumeric word that starts with alphabet, without containing space and punctuations.
// Underscore `_` can also be used within the key.
// String literal like "mickey mouse" cannot be used as a hash key.
// The internal key is actually a String and **not a Symbol** for now (TBD).
// Thus only a String object or a string literal should be used when referencing with `[ ]`.
//
// ```ruby
// a = { balthazar1: 100 } # valid
// b = { 2melchior: 200 }  # invalid
// x = 'balthazar1'
//
// a["balthazar1"]  # => 100
// a[x]             # => 100
// a[balthazar1]    # => error
// ```
//
// - **value:** String literal and objects (Integer, String, Array, Hash, nil, etc) can be used.
//
// **Note:**
// - The order of key-value pairs are **not** preserved.
// - Operator `=>` is not supported.
// - `Hash.new` is not supported.
type HashObject struct {
	*baseObj
	Pairs map[string]Object
}

// Class methods --------------------------------------------------------
func builtinHashClassMethods() []*BuiltinMethodObject {
	return []*BuiltinMethodObject{
		{
			Name: "new",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					return t.unsupportedMethodError("#new", receiver)
				}
			},
		},
	}
}

// Instance methods -----------------------------------------------------
func builtinHashInstanceMethods() []*BuiltinMethodObject {
	return []*BuiltinMethodObject{
		{
			// Retrieves the value (object) that corresponds to the key specified.
			// Returns `nil` when specifying a nonexistent key.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: 'v' } }
			// h['a'] #=> 1
			// h['b'] #=> "2"
			// h['c'] #=> [1, 2, 3]
			// h['d'] #=> { k: 'v' }
			// ```
			//
			// @return [Object]
			Name: "[]",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					if len(args) != 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 1 argument. got: %d", len(args))
					}

					i := args[0]
					key, ok := i.(*StringObject)

					if !ok {
						return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.StringClass, i.Class().Name)
					}

					h := receiver.(*HashObject)

					if len(h.Pairs) == 0 {
						return NULL
					}

					value, ok := h.Pairs[key.value]

					if !ok {
						return NULL
					}

					return value
				}
			},
		},
		{
			// Associates the value given by `value` with the key given by `key`.
			// Returns the `value`.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: 'v' } }
			// h['a'] = 1          #=> 1
			// h['b'] = "2"        #=> "2"
			// h['c'] = [1, 2, 3]  #=> [1, 2, 3]
			// h['d'] = { k: 'v' } #=> { k: 'v' }
			// ```
			//
			// @return [Object] The value
			Name: "[]=",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					// First arg is index
					// Second arg is assigned value
					if len(args) != 2 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 2 arguments. got: %d", len(args))
					}

					k := args[0]
					key, ok := k.(*StringObject)

					if !ok {
						return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.StringClass, k.Class().Name)
					}

					h := receiver.(*HashObject)
					h.Pairs[key.value] = args[1]

					return args[1]
				}
			},
		},
		{
			// Returns empty hash (no key-value pairs)
			//
			// ```Ruby
			// { a: "Hello", b: "World" }.clear # => {}
			// {}.clear                         # => {}
			// ```
			//
			// @return [Boolean]
			Name: "clear",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					return t.vm.initHashObject(make(map[string]Object))
				}
			},
		},
		{
			// Loop through keys of the hash with given block frame. It also returns array of
			// keys in alphabetical order.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: 'v' } }
			// h.each_key do |k|
			//   puts k
			// end
			// # => a
			// # => b
			// # => c
			// # => d
			// ```
			//
			// @return [Array]
			Name: "each_key",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					if blockFrame == nil {
						return t.vm.initErrorObject(errors.InternalError, errors.CantYieldWithoutBlockFormat)
					}

					h := receiver.(*HashObject)
					keys := h.sortedKeys()
					var arrOfKeys []Object

					for _, k := range keys {
						obj := t.vm.initStringObject(k)
						arrOfKeys = append(arrOfKeys, obj)
						t.builtinMethodYield(blockFrame, obj)
					}

					return t.vm.initArrayObject(arrOfKeys)
				}
			},
		},
		{
			// Loop through values of the hash with given block frame. It also returns array of
			// values of the hash in the alphabetical order of its key
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: "v" } }
			// h.each_value do |v|
			//   puts v
			// end
			// # => 1
			// # => "2"
			// # => [1, 2, 3]
			// # => { k: "v" }
			// ```
			//
			Name: "each_value",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					if blockFrame == nil {
						return t.vm.initErrorObject(errors.InternalError, errors.CantYieldWithoutBlockFormat)
					}

					h := receiver.(*HashObject)
					keys := h.sortedKeys()
					var arrOfValues []Object

					for _, k := range keys {
						value := h.Pairs[k]
						arrOfValues = append(arrOfValues, value)
						t.builtinMethodYield(blockFrame, value)
					}

					return t.vm.initArrayObject(arrOfValues)
				}
			},
		},
		{
			// Returns true if hash has no key-value pairs
			//
			// ```Ruby
			// {}.empty?       # => true
			// { a: 1 }.empty? # => false
			// ```
			//
			// @return [Boolean]
			Name: "empty?",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					if h.length() == 0 {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Returns true if hash is exactly equal to another hash
			//
			// ```Ruby
			// { a: "Hello", b: "World" }.eql?(1) # => false
			// ```
			//
			// @return [Boolean]
			Name: "eql?",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 1 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					c := args[0]
					compare, ok := c.(*HashObject)

					if ok && reflect.DeepEqual(h, compare) {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Remove the key from the hash if key exist
			//
			// ```Ruby
			// h = { a: 1, b: 2, c: 3 }
			// h.delete("b") # =>  { a: 1, c: 3 }
			// ```
			//
			// @return [Hash]
			Name: "delete",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 1 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					d := args[0]
					deleteKey, ok := d.(*StringObject)

					if !ok {
						return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.StringClass, d.Class().Name)
					}

					deleteKeyValue := deleteKey.value
					if _, ok := h.Pairs[deleteKeyValue]; ok {
						delete(h.Pairs, deleteKeyValue)
					}
					return h
				}
			},
		},
		{
			// Returns true if the key exist in the hash. Currently, it can only input string
			// type object.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: "v" } }
			// h.has_key?("a") # => true
			// h.has_key?("e") # => false
			// # TODO: Support Symbol Type Key Input
			// h.has_key?(:b)  # => true
			// h.has_key?(:f)  # => false
			// ```
			//
			// @return [Boolean]
			Name: "has_key?",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 1 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					i := args[0]
					input, ok := i.(*StringObject)

					if !ok {
						return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.StringClass, i.Class().Name)
					}

					if _, ok := h.Pairs[input.value]; ok {
						return TRUE
					}
					return FALSE
				}
			},
		},
		{
			// Returns true if the value exist in the hash.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: "v" } }
			// h.has_value?(1)          # => true
			// h.has_value?(2)          # => false
			// h.has_value?("2")        # => true
			// h.has_value?([1, 2, 3])  # => true
			// h.has_value?({ k: "v" }) # => true
			// ```
			//
			// @return [Boolean]
			Name: "has_value?",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 1 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)

					for _, v := range h.Pairs {
						if reflect.DeepEqual(v, args[0]) {
							return TRUE
						}
					}
					return FALSE
				}
			},
		},
		{
			// Returns an array of keys (in arbitrary order)
			//
			// ```Ruby
			// { a: 1, b: "2", c: [3, true, "Hello"] }.keys
			// # =>  ["c", "b", "a"] or ["b", "a", "c"] ... etc
			// ```
			//
			// @return [Boolean]
			Name: "keys",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					var keys []Object
					for k := range h.Pairs {
						keys = append(keys, t.vm.initStringObject(k))
					}
					return t.vm.initArrayObject(keys)
				}
			},
		},
		{
			// Returns the number of key-value pairs of the hash.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3], d: { k: 'v' } }
			// h.length  #=> 4
			// ```
			//
			// @return [Integer]
			Name: "length",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					return t.vm.initIntegerObject(h.length())
				}
			},
		},
		{
			// Returns a new hash with the results of running the block once for every value.
			// This method does not change the keys and the receiver hash values.
			//
			// ```Ruby
			// h = { a: 1, b: 2, c: 3 }
			// result = h.transform_values do |v|
			//   v * 3
			// end
			// h      # => { a: 3, b: 6, c: 9 }
			// result # => { a: 3, b: 6, c: 9 }
			// ```
			//
			// @return [Boolean]
			Name: "map_values",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					if blockFrame == nil {
						return t.vm.initErrorObject(errors.InternalError, errors.CantYieldWithoutBlockFormat)
					}

					h := receiver.(*HashObject)
					for k, v := range h.Pairs {
						result := t.builtinMethodYield(blockFrame, v)
						h.Pairs[k] = result.Target
					}
					return h
				}
			},
		},
		{
			// Returns the number of key-value pairs of the hash.
			//
			// ```Ruby
			// h = { a: 1, b: "2", c: [1, 2, 3] }
			// h.merge({ b: "Hello", d: "World" })
			// # => { a: 1, b: "Hello", c: [1, 2, 3], d: "World" }
			// ```
			//
			// @return [Hash]
			Name: "merge",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) < 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect at least 1 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					result := make(map[string]Object)
					for k, v := range h.Pairs {
						result[k] = v
					}

					for _, obj := range args {
						hashObj, ok := obj.(*HashObject)
						if !ok {
							return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.HashClass, obj.Class().Name)
						}
						for k, v := range hashObj.Pairs {
							result[k] = v
						}
					}

					return t.vm.initHashObject(result)
				}
			},
		},
		{
			// Returns an array of keys (in arbitrary order)
			//
			// ```Ruby
			// { a: 1, b: "2", c: [3, true, "Hello"] }.sorted_keys
			// # =>  ["a", "b", "c"]
			// { c: 1, b: "2", a: [3, true, "Hello"] }.sorted_keys
			// # =>  ["a", "b", "c"]
			// { b: 1, c: "2", a: [3, true, "Hello"] }.sorted_keys
			// # =>  ["a", "b", "c"]
			// { b: 1, c: "2", b: [3, true, "Hello"] }.sorted_keys
			// # =>  ["b", "c"]
			// ```
			//
			// @return [Boolean]
			Name: "sorted_keys",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					sortedKeys := h.sortedKeys()
					var keys []Object
					for _, k := range sortedKeys {
						keys = append(keys, t.vm.initStringObject(k))
					}
					return t.vm.initArrayObject(keys)
				}
			},
		},
		{
			// Returns two-dimensional array with the key-value pairs of hash. If specified true
			// then it will return sorted key value pairs array
			//
			// ```Ruby
			// { a: 1, b: 2, c: 3 }.to_a
			// # => [["a", 1], ["c", 3], ["b", 2]] or [["b", 2], ["c", 3], ["a", 1]] ... etc
			// { a: 1, b: 2, c: 3 }.to_a(true)
			// # => [["a", 1], ["b", 2], ["c", 3]]
			// { b: 1, a: 2, c: 3 }.to_a(true)
			// # => [["a", 2], ["b", 1], ["c", 3]]
			// { b: 1, a: 2, a: 3 }.to_a(true)
			// # => [["a", 3], ["b", 1]]
			// ```
			//
			// @return [Array]
			Name: "to_a",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {

					h := receiver.(*HashObject)
					var sorted bool

					if len(args) == 0 {
						sorted = false
					} else if len(args) > 1 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0..1 argument. got: %d", len(args))
					} else {
						s := args[0]
						st, ok := s.(*BooleanObject)
						if !ok {
							return t.vm.initErrorObject(errors.TypeError, errors.WrongArgumentTypeFormat, classes.BooleanClass, s.Class().Name)
						}
						sorted = st.value
					}

					var resultArr []Object
					if sorted {
						for _, k := range h.sortedKeys() {
							var pairArr []Object
							pairArr = append(pairArr, t.vm.initStringObject(k))
							pairArr = append(pairArr, h.Pairs[k])
							resultArr = append(resultArr, t.vm.initArrayObject(pairArr))
						}
					} else {
						for k, v := range h.Pairs {
							var pairArr []Object
							pairArr = append(pairArr, t.vm.initStringObject(k))
							pairArr = append(pairArr, v)
							resultArr = append(resultArr, t.vm.initArrayObject(pairArr))
						}
					}
					return t.vm.initArrayObject(resultArr)
				}
			},
		},
		{
			// Returns json that is corresponding to the hash.
			// Basically just like Hash#to_json in Rails but currently doesn't support options.
			//
			// ```Ruby
			// h = { a: 1, b: [1, "2", [4, 5, nil], { foo: "bar" }]}.to_json
			// puts(h) #=> {"a":1,"b":[1, "2", [4, 5, null], {"foo":"bar"}]}
			// ```
			//
			// @return [String]
			Name: "to_json",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					r := receiver.(*HashObject)
					return t.vm.initStringObject(r.toJSON())
				}
			},
		},
		{
			// Returns json that is corresponding to the hash.
			// Basically just like Hash#to_json in Rails but currently doesn't support options.
			//
			// ```Ruby
			// h = { a: 1, b: [1, "2", [4, 5, nil], { foo: "bar" }]}.to_s
			// puts(h) #=> "{ a: 1, b: [1, \"2\", [4, 5, null], { foo: \"bar \" }] }"
			// ```
			//
			// @return [String]
			Name: "to_s",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					return t.vm.initStringObject(h.toString())
				}
			},
		},
		{
			// Returns a new hash with the results of running the block once for every value.
			// This method does not change the keys and unlike Hash#map_values, it does not
			// change the receiver hash values.
			//
			// ```Ruby
			// h = { a: 1, b: 2, c: 3 }
			// result = h.transform_values do |v|
			//   v * 3
			// end
			// h      # => { a: 1, b: 2, c: 3 }
			// result # => { a: 3, b: 6, c: 9 }
			// ```
			//
			// @return [Boolean]
			Name: "transform_values",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					if blockFrame == nil {
						return t.vm.initErrorObject(errors.InternalError, errors.CantYieldWithoutBlockFormat)
					}

					h := receiver.(*HashObject)
					resultHash := make(map[string]Object)
					for k, v := range h.Pairs {
						result := t.builtinMethodYield(blockFrame, v)
						resultHash[k] = result.Target
					}
					return t.vm.initHashObject(resultHash)
				}
			},
		},
		{
			// Returns an array of values (in arbitrary order)
			//
			// ```Ruby
			// { a: 1, b: "2", c: [3, true, "Hello"] }.keys
			// # =>  [1, "2", [3, true, "Hello"]] or ["2", [3, true, "Hello"], 1] ... etc
			// ```
			//
			// @return [Boolean]
			Name: "values",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					if len(args) != 0 {
						return t.vm.initErrorObject(errors.ArgumentError, "Expect 0 argument. got: %d", len(args))
					}

					h := receiver.(*HashObject)
					var keys []Object
					for _, v := range h.Pairs {
						keys = append(keys, v)
					}
					return t.vm.initArrayObject(keys)
				}
			},
		},
	}
}

// Internal functions ===================================================

// Functions for initialization -----------------------------------------

func (vm *VM) initHashObject(pairs map[string]Object) *HashObject {
	return &HashObject{
		baseObj: &baseObj{class: vm.topLevelClass(classes.HashClass)},
		Pairs:   pairs,
	}
}

func (vm *VM) initHashClass() *RClass {
	hc := vm.initializeClass(classes.HashClass, false)
	hc.setBuiltinMethods(builtinHashInstanceMethods(), false)
	hc.setBuiltinMethods(builtinHashClassMethods(), true)
	return hc
}

// Polymorphic helper functions -----------------------------------------

// Returns the object
func (h *HashObject) Value() interface{} {
	return h.Pairs
}

// Returns the object's name as the string format
func (h *HashObject) toString() string {
	var out bytes.Buffer
	var pairs []string

	for _, key := range h.sortedKeys() {
		// TODO: Improve this conditional statement
		if _, isString := h.Pairs[key].(*StringObject); isString {
			pairs = append(pairs, fmt.Sprintf("%s: \"%s\"", key, h.Pairs[key].toString()))
		} else {
			pairs = append(pairs, fmt.Sprintf("%s: %s", key, h.Pairs[key].toString()))
		}
	}

	out.WriteString("{ ")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString(" }")

	return out.String()
}

// Returns the object's name as the JSON string format
func (h *HashObject) toJSON() string {
	var out bytes.Buffer
	var values []string
	pairs := h.Pairs
	out.WriteString("{")

	for key, value := range pairs {
		values = append(values, generateJSONFromPair(key, value))
	}

	out.WriteString(strings.Join(values, ","))
	out.WriteString("}")
	return out.String()
}

// Returns the length of the hash
func (h *HashObject) length() int {
	return len(h.Pairs)
}

// Returns the sorted keys of the hash
func (h *HashObject) sortedKeys() []string {
	var arr []string
	for k := range h.Pairs {
		arr = append(arr, k)
	}
	sort.Strings(arr)
	return arr
}

// Returns the duplicate of the Hash object
func (h *HashObject) copy() Object {
	elems := map[string]Object{}

	for k, v := range h.Pairs {
		elems[k] = v
	}

	newHash := &HashObject{
		baseObj: &baseObj{class: h.class},
		Pairs:   elems,
	}

	return newHash
}

// Other helper functions ----------------------------------------------

// Return the JSON style strings of the Hash object
func generateJSONFromPair(key string, v Object) string {
	var data string
	var out bytes.Buffer

	out.WriteString(data)
	out.WriteString("\"" + key + "\"")
	out.WriteString(":")
	out.WriteString(v.toJSON())

	return out.String()
}
