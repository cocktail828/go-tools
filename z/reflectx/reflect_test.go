package reflectx

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeCloser struct{}

func (FakeCloser) Close() error { return nil }

type SimpleStruct struct {
	Name   string
	Age    int
	Score  float64
	Active bool
}

type NestedStruct struct {
	ID     int
	Simple SimpleStruct
}

type StructWithUnexported struct {
	Public  string
	private string // unexported field
}

type StructWithPointer struct {
	ID    int
	Value *string
}

type CircularStruct struct {
	ID   int
	Next *CircularStruct
}

func TestIsNil(t *testing.T) {
	var k io.Closer = func() *FakeCloser {
		return nil
	}()

	assert.Equal(t, false, k == nil)
	assert.Equal(t, true, IsNil(k))
	assert.Equal(t, false, IsNil(net.AddrError{"test", "127.0.0.1"}))
}

func TestStringifyBasicTypes(t *testing.T) {
	// Test basic types
	assert.Contains(t, Stringify("hello"), "hello")
	assert.Contains(t, Stringify(42), "42")
	assert.Contains(t, Stringify(3.14), "3.14")
	assert.Contains(t, Stringify(true), "true")
	assert.Contains(t, Stringify(nil), "nil")
}

func TestStringifyStruct(t *testing.T) {
	// Test struct
	obj := SimpleStruct{
		Name:   "John",
		Age:    30,
		Score:  95.5,
		Active: true,
	}

	result := Stringify(obj)
	assert.Contains(t, result, "SimpleStruct")
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "Age")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "Score")
	assert.Contains(t, result, "95.5")
	assert.Contains(t, result, "Active")
	assert.Contains(t, result, "true")
}

func TestStringifyNestedStruct(t *testing.T) {
	// Test nested struct
	obj := NestedStruct{
		ID: 1,
		Simple: SimpleStruct{
			Name:   "Alice",
			Age:    25,
			Score:  90.0,
			Active: false,
		},
	}

	result := Stringify(obj)
	assert.Contains(t, result, "NestedStruct")
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "Simple")
	assert.Contains(t, result, "SimpleStruct")
	assert.Contains(t, result, "Alice")
}

func TestStringifyPointer(t *testing.T) {
	// Test pointer types
	str := "test string"
	ptr := &str

	result := Stringify(ptr)
	assert.Contains(t, result, "test string")

	// 测试nil指针
	var nilPtr *string
	assert.Contains(t, Stringify(nilPtr), "nil")

	// 测试结构体指针
	structPtr := &SimpleStruct{Name: "Bob", Age: 35}
	result = Stringify(structPtr)
	assert.Contains(t, result, "&SimpleStruct")
	assert.Contains(t, result, "Bob")
	assert.Contains(t, result, "35")
}

func TestStringifyStructWithPointer(t *testing.T) {
	// Test struct containing pointer
	str := "pointer value"
	obj := StructWithPointer{
		ID:    42,
		Value: &str,
	}

	result := Stringify(obj)
	assert.Contains(t, result, "StructWithPointer")
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "Value")
	assert.Contains(t, result, "pointer value")
}

func TestStringifyCircularReference(t *testing.T) {
	// Test circular reference
	node1 := &CircularStruct{ID: 1}
	node2 := &CircularStruct{ID: 2}
	node1.Next = node2
	node2.Next = node1

	result := Stringify(node1)
	assert.Contains(t, result, "CircularStruct")
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "<cycle detected>")
}

func TestStringifySlice(t *testing.T) {
	// Test slices
	slice := []string{"apple", "banana", "cherry"}
	result := Stringify(slice)
	assert.Contains(t, result, "string")
	assert.Contains(t, result, "apple")
	assert.Contains(t, result, "banana")
	assert.Contains(t, result, "cherry")

	// Test slice of structs
	structSlice := []SimpleStruct{
		{Name: "Tom", Age: 20},
		{Name: "Jerry", Age: 22},
	}
	result = Stringify(structSlice)
	assert.Contains(t, result, "SimpleStruct")
	assert.Contains(t, result, "Tom")
	assert.Contains(t, result, "Jerry")

	// Test nil slice
	var nilSlice []int
	assert.Contains(t, Stringify(nilSlice), "nil")
}

func TestStringifyArray(t *testing.T) {
	// Test arrays
	array := [3]int{1, 2, 3}
	result := Stringify(array)
	assert.Contains(t, result, "int")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "3")
}

func TestStringifyMap(t *testing.T) {
	// Test map with basic types
	strMap := map[string]int{"one": 1, "two": 2}
	result := Stringify(strMap)
	assert.Contains(t, result, "map")
	assert.Contains(t, result, "string")
	assert.Contains(t, result, "int")
	assert.Contains(t, result, "one")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "two")
	assert.Contains(t, result, "2")

	// Test map with structs
	structMap := map[string]SimpleStruct{
		"user1": {Name: "Admin", Age: 40},
		"user2": {Name: "User", Age: 30},
	}
	result = Stringify(structMap)
	assert.Contains(t, result, "map")
	assert.Contains(t, result, "string")
	assert.Contains(t, result, "SimpleStruct")
	assert.Contains(t, result, "user1")
	assert.Contains(t, result, "Admin")
	assert.Contains(t, result, "user2")
	assert.Contains(t, result, "User")

	// Test nil map
	var nilMap map[string]int
	assert.Contains(t, Stringify(nilMap), "nil")
}

func TestStringifyUnexportedFields(t *testing.T) {
	// Test unexported fields
	obj := StructWithUnexported{
		Public:  "visible",
		private: "hidden", // 不可导出字段应该被跳过
	}

	result := Stringify(obj)
	assert.Contains(t, result, "StructWithUnexported")
	assert.Contains(t, result, "Public")
	assert.Contains(t, result, "visible")
	// 不应该包含私有字段
	assert.NotContains(t, result, "private")
	assert.NotContains(t, result, "hidden")
}

func TestStringifyComplexNested(t *testing.T) {
	// Test complex nested structures
	str := "nested value"
	complexObj := struct {
		ID    int
		Items []map[string]*SimpleStruct
		Meta  map[string]interface{}
	}{
		ID: 100,
		Items: []map[string]*SimpleStruct{
			{
				"item1": {Name: "Item One", Age: 1},
			},
		},
		Meta: map[string]interface{}{
			"string": "text",
			"number": 42,
			"bool":   true,
			"ptr":    &str,
		},
	}

	result := Stringify(complexObj)
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "100")
	assert.Contains(t, result, "Items")
	assert.Contains(t, result, "item1")
	assert.Contains(t, result, "Item One")
	assert.Contains(t, result, "Meta")
	assert.Contains(t, result, "string")
	assert.Contains(t, result, "text")
	assert.Contains(t, result, "number")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "nested value")
}

// Define a struct with a stringify tag
type StructWithTag struct {
	PublicField  string
	HiddenField  string `stringify:"false"`
	AnotherField int
}

// Define a struct that implements Stringify()
type StructWithStringifyMethod struct {
	Name  string
	Value int
}

func (s StructWithStringifyMethod) Stringify() string {
	return fmt.Sprintf("CustomStringify{Name='%s', Value=%d}", s.Name, s.Value)
}

// Define a struct that contains a field which implements Stringify()
type StructWithCustomMethodField struct {
	ID   int
	Data StructWithStringifyMethod
}

func TestStringifyWithTag(t *testing.T) {
	// Test fields with stringify:"false" tag
	obj := StructWithTag{
		PublicField:  "visible",
		HiddenField:  "should be hidden",
		AnotherField: 42,
	}

	result := Stringify(obj)
	assert.Contains(t, result, "StructWithTag")
	assert.Contains(t, result, "PublicField")
	assert.Contains(t, result, "visible")
	assert.Contains(t, result, "AnotherField")
	assert.Contains(t, result, "42")
	// 不应该包含带有stringify:"false"标签的字段
	assert.NotContains(t, result, "HiddenField")
	assert.NotContains(t, result, "should be hidden")
}

func TestStringifyWithCustomMethod(t *testing.T) {
	// Test object that directly implements Stringify()
	obj := StructWithStringifyMethod{
		Name:  "TestObject",
		Value: 100,
	}

	result := Stringify(obj)
	// 应该使用自定义的Stringify()方法返回的值，而不是默认的结构体格式化
	assert.Contains(t, result, "CustomStringify{Name='TestObject', Value=100}")
	assert.NotContains(t, result, "StructWithStringifyMethod") // 不应该包含结构体名称
}

func TestStringifyWithCustomMethodField(t *testing.T) {
	// Test struct that contains a field implementing Stringify()
	obj := StructWithCustomMethodField{
		ID: 42,
		Data: StructWithStringifyMethod{
			Name:  "EmbeddedObject",
			Value: 200,
		},
	}

	result := Stringify(obj)
	assert.Contains(t, result, "StructWithCustomMethodField")
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "Data")
	// Data字段应该使用自定义的Stringify()方法返回的值
	assert.Contains(t, result, "CustomStringify{Name='EmbeddedObject', Value=200}")
}

func TestStringifyWithCustomMethodPtr(t *testing.T) {
	// Test pointer to a struct that implements Stringify()
	obj := &StructWithStringifyMethod{
		Name:  "PtrObject",
		Value: 300,
	}

	result := Stringify(obj)
	// 应该使用自定义的Stringify()方法返回的值
	assert.Contains(t, result, "CustomStringify{Name='PtrObject', Value=300}")
}

func TestStringifyAllFeaturesTogether(t *testing.T) {
	// Test all features together
	str := "nested string"
	complexObj := struct {
		ID          int `stringify:"false"`
		PublicData  string
		CustomData  StructWithStringifyMethod
		HiddenField *string `stringify:"false"`
		Items       []*StructWithTag
	}{
		ID:          999, // 应该被隐藏
		PublicData:  "This is visible",
		CustomData:  StructWithStringifyMethod{Name: "Complex", Value: 400},
		HiddenField: &str, // 应该被隐藏
		Items: []*StructWithTag{
			{PublicField: "Item1", HiddenField: "hidden1", AnotherField: 1},
			{PublicField: "Item2", HiddenField: "hidden2", AnotherField: 2},
		},
	}

	result := Stringify(complexObj)
	// verify visible fields
	assert.Contains(t, result, "PublicData")
	assert.Contains(t, result, "This is visible")
	assert.Contains(t, result, "CustomData")
	assert.Contains(t, result, "CustomStringify{Name='Complex', Value=400}")
	assert.Contains(t, result, "Items")
	assert.Contains(t, result, "Item1")
	assert.Contains(t, result, "Item2")
	assert.Contains(t, result, "AnotherField")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")

	// varify hidden fields
	assert.NotContains(t, result, "ID")
	assert.NotContains(t, result, "999")
	assert.NotContains(t, result, "HiddenField")
	assert.NotContains(t, result, "nested string")
	assert.NotContains(t, result, "hidden1")
	assert.NotContains(t, result, "hidden2")
}
