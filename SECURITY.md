# Security Policy

## Supported Versions

We actively monitor and provide security updates for the following versions of **OpenCode-Native**:

| Version | Supported          |
| ------- | ------------------ |
| v1.x    | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.** 

We take the security of our users and their data seriously. If you believe you have found a security vulnerability, please help us fix it by reporting it responsibly.

### How to report
To report a vulnerability, please use one of the following methods:

1.  **GitHub Private Vulnerability Reporting:** Navigate to the [Security Tab](https://github.com/MerrukTechnology/OpenCode-Native/security/advisories/new) of this repository and select **"Report a vulnerability"**. This is the preferred and most secure method.
2.  **Email:** Send an encrypted report to **security@merruk.com**.

### 🔒 Encrypted Reports
For sensitive information, please encrypt your report using our PGP public key. 

- **Public Key:** [.github/SECURITY_PUBKEY.asc](https://github.com/MerrukTechnology/OpenCode-Native/raw/main/.github/SECURITY_PUBKEY.asc)
- **PGP Fingerprint:** `9A19C382B6616F7CB371797DCD2C6694D3274AB2`

You can find the public key attached to the GitHub profile of [@MerrukTechnology](https://github.com/MerrukTechnology) or by running:
```bash
gpg --keyserver keyserver.ubuntu.com --recv-keys 9A19C382B6616F7CB371797DCD2C6694D3274AB2
```

### What to include
To help us triage the issue quickly, please include:
*   A description of the vulnerability and its potential impact.
*   Step-by-step instructions to reproduce the issue (a Proof of Concept).
*   Any suggested fixes or mitigations.

## Our Response Process

After receiving a report, the **MerrukTechnology** security team will:
1.  Acknowledge receipt of the report within **48 hours**.
2.  Conduct an investigation to verify the vulnerability.
3.  Provide an estimated timeline for a fix.
4.  Once the fix is ready, we will release a new version and notify you.

## Disclosure Policy

We follow the principle of **Coordinated Vulnerability Disclosure**. We ask that you do not share information about the vulnerability publicly until we have had a reasonable amount of time to push a fix and notify affected users.

## Security Tools in Use

This project is continuously scanned for vulnerabilities using:
*   **CodeQL:** For deep semantic analysis and common CWE detection.
*   **GolangCI-Lint:** For Go-specific security (gosec) and code quality.
*   **Codacy:** For automated security reviews and standards compliance.

---
*Thank you for helping keep OpenCode-Native secure!*
