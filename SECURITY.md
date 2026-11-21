# Security Policy

## Reporting a vulnerability

Please do not report security vulnerabilities in public issues. Use GitHub's private
vulnerability reporting feature, if enabled, or contact the repository maintainers privately with
the affected component, reproduction steps, and potential impact.

Do not include real patient data, credentials, tokens, or other sensitive information in a report.

## Security expectations

- Never commit `.env` files, credentials, JWT secrets, or production database URLs.
- Keep account passwords and public-access passwords separate and hashed.
- Do not expose password hashes or clinical data through logs or error responses.
- All changes should pass the required CI, dependency, vulnerability, and CodeQL checks before
  being merged into `main`.
