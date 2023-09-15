package dotenv

import (
	"bytes"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	env := Environment{
		"TEST1": "test1",
		"TEST2": "test2",
	}

	var envBuffer bytes.Buffer
	for key, value := range env {
		envBuffer.WriteString(key)
		envBuffer.WriteRune('=')
		envBuffer.WriteString(value)
		envBuffer.WriteRune('\n')
	}

	file, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(envBuffer.Bytes()); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err.Error())
	}

	oldEnv := Load(file)
	defer oldEnv.Restore()

	for key, value := range env {
		if envValue := os.Getenv(key); envValue != value {
			t.Errorf("[Test] Invalid environment: received %s=%s expected %s=%s", key, envValue, key, value)
		}
	}
}

func TestOldEnv(t *testing.T) {
	env := Environment{
		"TEST1": "test1",
		"TEST2": "test2",
	}

	var envBuffer bytes.Buffer
	for key, value := range env {
		os.Setenv(key, value)
		envBuffer.WriteString(key)
		envBuffer.WriteRune('=')
		envBuffer.WriteString(value)
		envBuffer.WriteRune('\n')
	}

	file, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(envBuffer.Bytes()); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err.Error())
	}

	oldEnv := Load(file)
	defer oldEnv.Restore()

	for key, value := range env {
		if value != (*oldEnv)[key] {
			t.Errorf("[Test] Invalid old environment: received %s=%s expected %s=%s", key, (*oldEnv)[key], key, value)
		}
	}
}

func TestComment(t *testing.T) {
	env := Environment{
		"TEST1":   "test1",
		"TEST2":   "test2",
		"# TEST3": "test2",
	}

	var envBuffer bytes.Buffer
	for key, value := range env {
		envBuffer.WriteString(key)
		envBuffer.WriteRune('=')
		envBuffer.WriteString(value)
		envBuffer.WriteRune('\n')
	}

	file, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(envBuffer.Bytes()); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err.Error())
	}

	oldEnv := Load(file)
	defer oldEnv.Restore()

	for key, value := range env {
		if key[0] == COMMENT {
			if os.Getenv(key) == value {
				t.Errorf("[Test] Invalid environment comment: %s=%s", key, value)
			}
		} else if envValue := os.Getenv(key); envValue != value {
			t.Errorf("[Test] Invalid environment: received %s=%s expected %s=%s", key, envValue, key, value)
		}
	}
}

func TestRestore(t *testing.T) {
	oldEnvKey := "TEST1"
	oldEnvValue := "test"
	os.Setenv(oldEnvKey, oldEnvValue)

	env := Environment{
		oldEnvKey: "test1",
		"TEST2":   "test2",
	}

	var envBuffer bytes.Buffer
	for key, value := range env {
		envBuffer.WriteString(key)
		envBuffer.WriteRune('=')
		envBuffer.WriteString(value)
		envBuffer.WriteRune('\n')
	}

	file, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(envBuffer.Bytes()); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err.Error())
	}

	oldEnv := Load(file)

	if envValue := os.Getenv(oldEnvKey); envValue != env[oldEnvKey] {
		t.Fatalf("[Test] Invalid old environment value: received %s=%s expected %s=%s", oldEnvKey, envValue, oldEnvKey, env[oldEnvKey])
	}

	oldEnv.Restore()

	if envValue := os.Getenv(oldEnvKey); envValue != oldEnvValue {
		t.Fatalf("[Test] Invalid old environment value: received %s=%s expected %s=%s", oldEnvKey, envValue, oldEnvKey, oldEnvValue)
	}
}
