# Error Budget and Reliability Policy

## Service Level Objective (SLO)
For staging reliability gates, Nduracore API targets:
- Availability: 99.5% over rolling 30 days.
- Successful request ratio: at least 98% non-5xx responses.

## Error Budget
- Monthly budget at 99.5% availability is approximately 3h 36m downtime.
- Budget burn is monitored weekly in staging and continuously in production.

## Policy
1. If error budget burn exceeds 50% in the first half of the window:
   - Freeze non-critical feature work.
   - Prioritize reliability fixes.
2. If burn exceeds 100%:
   - Stop feature releases until SLO compliance is restored.
3. Every major incident must produce:
   - Incident summary.
   - Root cause analysis.
   - Action items with owner and due date.

## Core Alerting Signals
- `readiness` failing.
- Elevated 5xx rates.
- Latency regression on key endpoints.
- Authentication failure anomaly.

## Reporting Cadence
- Weekly reliability review in staging.
- Monthly SLO and error-budget review.
