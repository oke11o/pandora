package str

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStringFunc(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		input         string
		expectedName  string
		expectedArgs  []string
		expectedError error
	}{
		// Test case 1: Valid input without arguments
		{
			name:          "TestValidInputNoArgs",
			input:         "functionName",
			expectedName:  "functionName",
			expectedArgs:  nil,
			expectedError: nil,
		},
		// Test case 2: Valid input with arguments
		{
			name:          "TestValidInputWithArgs",
			input:         "functionName(arg1, arg2, arg3)",
			expectedName:  "functionName",
			expectedArgs:  []string{"arg1", "arg2", "arg3"},
			expectedError: nil,
		},
		// Test case 3: Invalid close bracket position
		{
			name:          "TestInvalidCloseBracket",
			input:         "functionName(arg1, arg2, arg3",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: errors.New("invalid close bracket position"),
		},
		// Test case 4: Valid input with one argument
		{
			name:          "TestValidInputOneArg",
			input:         "functionName(arg1)",
			expectedName:  "functionName",
			expectedArgs:  []string{"arg1"},
			expectedError: nil,
		},
		// Test case 5: Empty input
		{
			name:          "TestEmptyInput",
			input:         "",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: nil,
		},
		// Test case 6: Input with only open bracket
		{
			name:          "TestOnlyOpenBracket",
			input:         "(",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: errors.New("invalid close bracket position"),
		},
		// Test case 7: Input with only close bracket
		{
			name:          "TestOnlyCloseBracket",
			input:         ")",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: errors.New("invalid close bracket position"),
		},
		// Test case 8: Input with a single empty argument
		{
			name:          "TestSingleEmptyArgument",
			input:         "functionName()",
			expectedName:  "functionName",
			expectedArgs:  []string{""},
			expectedError: nil,
		},
		// Test case 9: Input with ')' as part of the function name
		{
			name:          "TestBracketInFunctionName",
			input:         "functionName)arg1, arg2, arg3)",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: errors.New("invalid close bracket position"),
		},
		// Test case 10: Input with ')' after the closing bracket
		{
			name:          "TestExtraCloseBracket",
			input:         "functionName(arg1, arg2, arg3))",
			expectedName:  "",
			expectedArgs:  nil,
			expectedError: errors.New("invalid close bracket position"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name, args, err := ParseStringFunc(tc.input)

			// Assert the values
			assert.Equal(t, tc.expectedName, name)
			assert.Equal(t, tc.expectedArgs, args)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
