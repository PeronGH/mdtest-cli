# Time Minute Parity Smoke Test

## Steps

1. Read the current local system time from this device.
2. Extract the current minute value as an integer from 0 to 59.
3. If the minute value is even, set status to `pass`.
4. If the minute value is odd, set status to `fail`.
5. In the log body, include the observed local time and the minute value that was used.
