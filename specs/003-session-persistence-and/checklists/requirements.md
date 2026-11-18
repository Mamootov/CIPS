# Requirements Quality Checklist: 003 - Session Persistence & Accuracy

- [x] **Zero-Assumption Empirical Ground Truth**: No static GeoIP or country blacklists are applied as pre-filters. Google's own localized verdict via Anycast is evaluated dynamically per node.
- [x] **No Premature Code Gates**: Requirements are grounded in user-verified observations and reproducible edge cases.
- [x] **Persistence Verification**: Client state is hydrated upon page reload via dedicated REST endpoint.
- [x] **Real-Time Data Availability**: Subscription generation reflects current clean nodes mid-scan without requiring session completion.
- [x] **Failure Isolation**: Batch startup failures are trapped and routed to individual single-instance runners.
- [x] **Socket Polling Budget**: Poller timeout is calibrated to Windows host process spawning latency (5000ms).
- [x] **UI Completeness**: No matrix cells are left in an indeterminate or blank state on completed evaluations.
