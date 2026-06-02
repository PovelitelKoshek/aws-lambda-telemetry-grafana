Errors over 5 minutes

```
increase(lambda_errors_total[5m])
```

Average duration

```
rate(lambda_duration_ms_sum[5m]) / rate(lambda_duration_ms_count[5m])
```

Error rate
```
increase(lambda_errors_total[5m]) / increase(lambda_invocations_total[5m]) * 100
```
