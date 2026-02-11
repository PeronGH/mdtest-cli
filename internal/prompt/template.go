package prompt

import "fmt"

func Render(testAbs string, logAbs string) string {
	return fmt.Sprintf(
		"Execute the test file step by step.\nRead the test from this exact absolute path: %s\nWrite the output log to this exact absolute path: %s\nThe output must begin with YAML front matter containing status: pass|fail.\n",
		testAbs,
		logAbs,
	)
}
