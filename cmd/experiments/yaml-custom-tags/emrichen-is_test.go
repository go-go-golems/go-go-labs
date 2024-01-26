package main

import "testing"

func TestIsTag_IsBoolean(t *testing.T) {
	tests := []testCase{
		{name: "IsBoolean with valid boolean", inputYAML: "!IsBoolean true", expected: "true"},
		{name: "IsBoolean with invalid boolean", inputYAML: "!IsBoolean \"not a boolean\"", expected: "false"},
		{name: "IsBoolean with null", inputYAML: "!IsBoolean null", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsDict(t *testing.T) {
	tests := []testCase{
		{name: "IsDict with valid dict", inputYAML: "!IsDict {key: \"value\"}", expected: "true"},
		{name: "IsDict with invalid dict", inputYAML: "!IsDict \"not a dict\"", expected: "false"},
		{name: "IsDict with null", inputYAML: "!IsDict null", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsInteger(t *testing.T) {
	tests := []testCase{
		{name: "IsInteger with valid integer", inputYAML: "!IsInteger 42", expected: "true"},
		{name: "IsInteger with invalid integer", inputYAML: "!IsInteger 3.14", expected: "false"},
		{name: "IsInteger with null", inputYAML: "!IsInteger null", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsList(t *testing.T) {
	tests := []testCase{
		{name: "IsList with valid list", inputYAML: "!IsList [1, 2, 3]", expected: "true"},
		{name: "IsList with invalid list", inputYAML: "!IsList {key: \"value\"}", expected: "false"},
		{name: "IsList with null", inputYAML: "!IsList null", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsNone(t *testing.T) {
	tests := []testCase{
		{name: "IsNone with valid none", inputYAML: "!IsNone null", expected: "true"},
		{name: "IsNone with invalid none", inputYAML: "!IsNone \"not null\"", expected: "false"},
		{name: "IsNone with 0", inputYAML: "!IsNone 0", expected: "false"},
		{name: "IsNone with 1", inputYAML: "!IsNone 1", expected: "false"},
		{name: "IsNone with empty string", inputYAML: "!IsNone \"\"", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsNumber(t *testing.T) {
	tests := []testCase{
		{name: "IsNumber with valid number", inputYAML: "!IsNumber 3.14", expected: "true"},
		{name: "IsNumber with invalid number", inputYAML: "!IsNumber \"not a number\"", expected: "false"},
		{name: "IsNumber with null", inputYAML: "!IsNumber null", expected: "false"},
	}

	runTests(t, tests)
}

func TestIsTag_IsString(t *testing.T) {
	tests := []testCase{
		{name: "IsString with valid string", inputYAML: "!IsString \"a string\"", expected: "true"},
		{name: "IsString with invalid string", inputYAML: "!IsString []", expected: "false"},
	}

	runTests(t, tests)
}
