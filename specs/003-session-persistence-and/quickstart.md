# Quickstart: 003 - Session Persistence & Ground Truth Accuracy

---

## 1. Verifying Session Hydration Across Refreshes

1. Ingest a list of proxy configurations into the dashboard at `http://127.0.0.1:19090`.
2. Start the scan or allow it to complete.
3. Refresh the browser tab using **F5** or `Ctrl+R`.
4. **Verification**: Observe that all matrix rows, latencies, ISP tags, and active status counters are fully restored immediately without data loss.

---

## 2. Copying Verified Nodes Mid-Scan

1. Start scanning a batch of proxies.
2. While the scan is still running (`Probing AI...` state), click **📋 Copy Clean Subscription**.
3. **Verification**: If any nodes have completed and verified as clean so far, they are formatted with smart tags and copied directly to the clipboard.

---

## 3. Purging Configurations via Clear All

1. In the sidebar under **3. Export & Actions**, click **🗑️ Clear All Configs**.
2. Confirm the prompt dialog.
3. **Verification**: The matrix grid resets to the initial empty state, and any in-progress scans are cancelled.
