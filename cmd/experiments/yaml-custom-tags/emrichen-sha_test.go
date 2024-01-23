package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Helper function to generate an MD5 hash
func md5Hash(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Helper function to generate a SHA1 hash
func sha1Hash(data string) string {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Helper function to generate a SHA256 hash
func sha256Hash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func TestEmrichenMD5Tag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic MD5 hashing",
			inputYAML: `!MD5 'hello world'`,
			expected:  md5Hash("hello world"),
		},
		{
			name:      "MD5 hashing empty string",
			inputYAML: `!MD5 ''`,
			expected:  md5Hash(""),
		},
		{
			name:      "MD5 hashing special characters",
			inputYAML: `!MD5 'special!@#$%'`,
			expected:  md5Hash("special!@#$%"),
		},
		{
			name:      "Encode numeric value",
			inputYAML: `!MD5 42`,
			expected:  md5Hash("42"),
		},
		{
			name:      "Encode boolean value",
			inputYAML: `!MD5 true`,
			expected:  md5Hash("true"),
		},
	}

	runTests(t, tests)
}

func TestEmrichenSHA1Tag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic SHA1 hashing",
			inputYAML: `!SHA1 'hello world'`,
			expected:  sha1Hash("hello world"),
		},
		{
			name:      "SHA1 hashing empty string",
			inputYAML: `!SHA1 ''`,
			expected:  sha1Hash(""),
		},
		{
			name:      "SHA1 hashing special characters",
			inputYAML: `!SHA1 'special!@#$%'`,
			expected:  sha1Hash("special!@#$%"),
		},
		{
			name:      "SHA1 hashing special characters",
			inputYAML: `!SHA1 'special!@#$%'`,
			expected:  sha1Hash("special!@#$%"),
		},
		{
			name:      "Encode numeric value",
			inputYAML: `!SHA1 42`,
			expected:  sha1Hash("42"),
		},
		{
			name:      "Encode boolean value",
			inputYAML: `!SHA1 true`,
			expected:  sha1Hash("true"),
		},
	}

	runTests(t, tests)
}

func TestEmrichenSHA256Tag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic SHA256 hashing",
			inputYAML: `!SHA256 'hello world'`,
			expected:  sha256Hash("hello world"),
		},
		{
			name:      "SHA256 hashing empty string",
			inputYAML: `!SHA256 ''`,
			expected:  sha256Hash(""),
		},
		{
			name:      "SHA256 hashing special characters",
			inputYAML: `!SHA256 'special!@#$%'`,
			expected:  sha256Hash("special!@#$%"),
		},
		{
			name:      "SHA256 hashing special characters",
			inputYAML: `!SHA256 'special!@#$%'`,
			expected:  sha256Hash("special!@#$%"),
		},
		{
			name:      "Encode numeric value",
			inputYAML: `!SHA256 42`,
			expected:  sha256Hash("42"),
		},
		{
			name:      "Encode boolean value",
			inputYAML: `!SHA256 true`,
			expected:  sha256Hash("true"),
		},
	}

	runTests(t, tests)
}

func TestEmrichenBase64Tag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic Base64 encoding",
			inputYAML: `!Base64 'hello world'`,
			expected:  "aGVsbG8gd29ybGQ=",
		},
		{
			name:      "Base64 encoding empty string",
			inputYAML: `!Base64 ''`,
			expected:  `""`,
		},
		{
			name:      "Base64 encoding special characters",
			inputYAML: `!Base64 'special!@#$%'`,
			expected:  "c3BlY2lhbCFAIyQl",
		},
		{
			name:      "Encode numeric value",
			inputYAML: `!Base64 42`,
			expected:  "NDI=",
		},
		{
			name:      "Encode boolean value",
			inputYAML: `!Base64 true`,
			expected:  "dHJ1ZQ==",
		},
	}

	runTests(t, tests)
}
