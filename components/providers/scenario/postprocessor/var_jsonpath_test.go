package postprocessor

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarJsonpathPostprocessor_Process(t *testing.T) {
	testCases := []struct {
		name      string
		mappings  map[string]string
		body      []byte
		expected  map[string]interface{}
		expectErr bool
	}{
		{
			name: "Test Case 1",
			mappings: map[string]string{
				"person_name": "$.name",
				"person_age":  "$.age",
			},
			body: []byte(`{"name": "John", "age": 30}`),
			expected: map[string]interface{}{
				"person_name": "John",
				"person_age":  float64(30),
			},
			expectErr: false,
		},
		{
			name: "Test Case 2",
			mappings: map[string]string{
				"user_name": "$.username",
				"user_age":  "$.age",
			},
			body: []byte(`{"username": "Alice", "age": 25}`),
			expected: map[string]interface{}{
				"user_name": "Alice",
				"user_age":  float64(25),
			},
			expectErr: false,
		},
		{
			name: "Test Case 3 - JSON parsing error",
			mappings: map[string]string{
				"name": "$.name",
			},
			body:      []byte(`invalid json`),
			expected:  map[string]interface{}{},
			expectErr: true,
		},
		{
			name: "Test Case 4 - Missing JSON field",
			mappings: map[string]string{
				"address": "$.address",
			},
			body:      []byte(`{"name": "Bob", "age": 35}`),
			expected:  map[string]interface{}{},
			expectErr: true,
		},
		{
			name: "Test Case 5 - Nested JSON",
			mappings: map[string]string{
				"city":      "$.address.city",
				"zip_code":  "$.address.zip",
				"country":   "$.address.country",
				"full_name": "$.personal.name.full",
			},
			body: []byte(`{
				"personal": {
					"name": {
						"first": "Jane",
						"last": "Doe",
						"full": "Jane Doe"
					},
					"age": 28
				},
				"address": {
					"city": "New York",
					"zip": "10001",
					"country": "USA"
				}
			}`),
			expected: map[string]interface{}{
				"city":      "New York",
				"zip_code":  "10001",
				"country":   "USA",
				"full_name": "Jane Doe",
			},
			expectErr: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a VarJsonpathPostprocessor instance with the mappings
			p := &VarJsonpathPostprocessor{Mapping: tc.mappings}

			// Initialize reqMap with some sample data for testing
			reqMap := make(map[string]interface{})

			// Call the Process method with the sample body
			err := p.Process(reqMap, &http.Response{}, tc.body)

			// Check if an error is expected and if it matches the actual result
			if tc.expectErr {
				assert.Error(t, err, "Expected an error, but got none")
			} else {
				assert.NoError(t, err, "Process should not return an error")
			}

			// Check if the reqMap is updated as expected
			assert.Equal(t, tc.expected, reqMap, "Process result not as expected")
		})
	}
}
