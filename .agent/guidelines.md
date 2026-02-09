# Project Guidelines v2026.02

## 1. Coding Standards
- **Clarity over Cleverness**: Write code that is easy to read and understand.
- **Comments**: Explain *why*, not *what*.
- **Error Handling**: Fail gracefully and log errors with context.

## 2. Security Standards
- **Credentials**: Never hardcode. Use `.env` or `_secrets.yaml`.
- **GPG Signing (Remote Only)**: All commits to remote repositories must be signed.
- **Output**: Mask sensitive data in logs/console output.

## 3. Workflow Standards
- **Unattended Execution**: Scripts should be runnable without interaction where possible (use flags).
- **Idempotency**: Scripts should be safe to run multiple times.
- **Documentation**: Update `README.md` or `workflows/` when changing processes.
