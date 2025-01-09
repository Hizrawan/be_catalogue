package reftype

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIsNil(t *testing.T) {
	type TestStruct struct {
		Data string
	}

	t.Run("returns true if provided value is a nil", func(t *testing.T) {
		var testArray []string
		var testPtr *string

		cases := []struct {
			Value any
		}{
			{nil},
			{testArray},
			{testPtr},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("testing case %d", i), func(t *testing.T) {
				res := IsNil(c.Value)
				if res != true {
					t.Errorf("want %v; got %v", true, res)
				}
			})
		}
	})

	t.Run("returns false if provided value is a not nil", func(t *testing.T) {
		testStruct := TestStruct{}
		testString := "chaldea"
		testInt := 3
		testArray := make([]string, 0)
		testPtr := &testInt
		var emptyStruct TestStruct
		var emptyString string
		var emptyInt int

		cases := []struct {
			Value any
		}{
			{testStruct},
			{testString},
			{testInt},
			{testArray},
			{testPtr},
			{emptyStruct},
			{emptyString},
			{emptyInt},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("testing case %d", i), func(t *testing.T) {
				res := IsNil(c.Value)
				if res != false {
					t.Errorf("want %v; got %v", false, res)
				}
			})
		}
	})
}

func TestIsTypeOf(t *testing.T) {
	type TestStructA struct {
		Data string
	}

	type TestStructB struct {
		Data string
	}

	var (
		testStructAIns = TestStructA{Data: "chaldea"}
		testStructBIns = TestStructB{Data: "rhodes island"}
	)

	t.Run("returns true if type is identical", func(t *testing.T) {
		cases := []struct {
			Match any
			With  any
		}{
			{true, false},
			{"This is", "That is"},
			{1, 2},
			{3.141592, 6.283184},
			{testStructAIns, TestStructA{Data: "artoria caster"}},
			{TestStructB{Data: "data1"}, TestStructB{Data: "data2"}},
			{&testStructBIns, &TestStructB{Data: "pointer"}},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("testing case %d", i), func(t *testing.T) {
				identical := IsTypeOf(c.Match, c.With)
				if !identical {
					t.Errorf("want %v; got %v", true, identical)
				}
			})
		}
	})

	t.Run("returns false if type is not identical", func(t *testing.T) {
		cases := []struct {
			Match any
			With  any
		}{
			{true, 1},
			{"This is", true},
			{1, "str"},
			{3.141592, 2},
			{[]string{}, []int{}},
			{testStructAIns, testStructBIns},
			{TestStructB{Data: "data1"}, TestStructA{Data: "data2"}},
			{&testStructAIns, testStructBIns},
			{[]string{}, []int{}},
			{&testStructAIns, &testStructBIns},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("testing case %d", i), func(t *testing.T) {
				identical := IsTypeOf(c.Match, c.With)
				if identical {
					t.Errorf("want %v; got %v", false, identical)
				}
			})
		}
	})

	t.Run("performs correctly with some elements being reflect.Type", func(t *testing.T) {
		cases := []struct {
			Match any
			With  any
			Want  bool
		}{
			{1, reflect.TypeOf(2), true},
			{testStructAIns, reflect.TypeOf(TestStructA{Data: "artoria caster"}), true},
			{reflect.TypeOf(TestStructB{Data: "data1"}), TestStructB{Data: "data2"}, true},
			{&testStructBIns, reflect.TypeOf(&TestStructB{Data: "pointer"}), true},
			{reflect.TypeOf(testStructAIns), testStructAIns, true},
			{reflect.TypeOf(testStructAIns), reflect.TypeOf(testStructAIns), true},
			{&testStructBIns, reflect.TypeOf(&TestStructA{Data: "pointer"}), false},
			{TestStructB{Data: "data1"}, TestStructA{Data: "data2"}, false},
			{&testStructAIns, testStructBIns, false},
			{[]string{}, reflect.TypeOf([]int{}), false},
			{reflect.TypeOf(&testStructAIns), reflect.TypeOf(&testStructBIns), false},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("testing case %d", i), func(t *testing.T) {
				identical := IsTypeOf(c.Match, c.With)
				if identical != c.Want {
					t.Errorf("want %v; got %v", c.Want, identical)
				}
			})
		}
	})
}

func TestIsStructEmbeds(t *testing.T) {
	type File struct {
		Filename string
	}

	type UpgradedFile struct {
		File
	}

	t.Run("returns true if the provided struct embeds the other", func(t *testing.T) {
		ins := UpgradedFile{}
		ref := File{}
		embeds := IsStructEmbeds(ins, ref)
		if !embeds {
			t.Errorf("want %v; got %v", true, embeds)
		}
	})

	type DifferentStruct struct {
		Size int
	}

	type SimilarStruct struct {
		Filename string
	}

	type AsAttribute struct {
		File File
	}

	type NotDirectEmbed struct {
		UpgradedFile
	}

	t.Run("returns false if the provided struct doesn't embed the other", func(t *testing.T) {
		cases := []struct {
			Name string
			Data any
		}{
			{"entirely different struct", DifferentStruct{}},
			{"similar struct", SimilarStruct{}},
			{"added as attribute", AsAttribute{}},
			{"not directly embedded", NotDirectEmbed{}},
		}
		ref := File{}

		for _, c := range cases {
			t.Run(fmt.Sprintf("testing case %s", c.Name), func(t *testing.T) {
				embed := IsStructEmbeds(c.Data, ref)
				if embed {
					t.Errorf("want %v; got %v", false, embed)
				}
			})
		}
	})

	t.Run("works properly even if some value is a reflect.Type instance", func(t *testing.T) {
		ref := File{}
		cases := []struct {
			Name string
			Ref  any
			Data any
			Want bool
		}{
			{"identical struct, comparison is a reflected type", ref, reflect.TypeOf(UpgradedFile{}), true},
			{"identical struct, both are reflected types", reflect.TypeOf(ref), reflect.TypeOf(UpgradedFile{}), true},
			{"entirely different struct, ref is a reflected type", reflect.TypeOf(ref), DifferentStruct{}, false},
			{"entirely different struct, both are reflected types", reflect.TypeOf(ref), reflect.TypeOf(DifferentStruct{}), false},
			{"similar struct, comparison is a reflected type", ref, reflect.TypeOf(SimilarStruct{}), false},
			{"similar struct, both are reflected types", reflect.TypeOf(ref), reflect.TypeOf(SimilarStruct{}), false},
			{"added as attribute, ref is a reflected type", reflect.TypeOf(ref), AsAttribute{}, false},
			{"added as attribute, both are reflected types", reflect.TypeOf(ref), reflect.TypeOf(AsAttribute{}), false},
			{"not directly embedded, comparison is a reflected type", ref, reflect.TypeOf(NotDirectEmbed{}), false},
			{"not directly embedded, both are reflected types", reflect.TypeOf(ref), reflect.TypeOf(NotDirectEmbed{}), false},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("testing case %s", c.Name), func(t *testing.T) {
				embed := IsStructEmbeds(c.Data, ref)
				if embed != c.Want {
					t.Errorf("want %v; got %v", c.Want, embed)
				}
			})
		}
	})
}
