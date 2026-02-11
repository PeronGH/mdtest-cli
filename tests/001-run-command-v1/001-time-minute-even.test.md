# Time Minute Parity Smoke Test

Goal: validate that `mdtest run` can execute a Markdown test and consume a real log result from the local device.

## Steps

1. Read the current local system time from this device.
2. Extract the current minute value as an integer from 0 to 59.
3. If the minute value is even, write a result log with YAML front matter `status: pass`.
4. If the minute value is odd, write a result log with YAML front matter `status: fail`.
5. In the log body, include the observed local time and the minute value that was used.
