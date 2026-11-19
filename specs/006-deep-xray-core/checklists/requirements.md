# Requirements Quality Checklist: 006 - Deep Xray Core Diagnostic Logging

- [x] **Zero-Assumption Empirical Ground Truth**: Errors are extracted directly from official Xray core stderr runtime streams, not guessed.
- [x] **Granular Classification**: Common failure signatures (Bad Reality Key, Bad Cipher, Bad UUID, Remote Refused) have distinct badges and descriptions.
- [x] **Modal Transparency**: The raw stderr log snippet is exposed in the Diagnostic Inspection Modal.
- [x] **Bounded Memory**: Logs are capped to the most recent 3 lines per node to prevent buffer bloat.
- [x] **Non-Negotiable**: No commit without user request; no git push without explicit user command.
