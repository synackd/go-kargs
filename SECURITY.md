<!--
SPDX-FileCopyrightText: © 2026 synack.d

SPDX-License-Identifier: BSD-3-Clause
-->

# Security Policy

## Reporting a vulnerability

Please report suspected security vulnerabilities in go-kargs through
[GitHub's private vulnerability reporting](https://github.com/synackd/go-kargs/security/advisories/new).
You can also open the repository's **Security** tab, select **Advisories**, and
click **Report a vulnerability**. A GitHub account is required to submit a report.

Do not disclose vulnerability details in public issues, pull requests, or
discussions. Use the private report for follow-up information and proposed fixes.
For ordinary bugs and feature requests, use
[GitHub issues](https://github.com/synackd/go-kargs/issues).

## What to include

To help maintainers reproduce and assess the issue, include:

- The affected go-kargs version or commit, Go version, and relevant environment.
- A description of the vulnerability, its potential impact, and any conditions
  required to exploit it.
- Steps to reproduce, preferably with a minimal Go example and the kernel command
  line input that triggers the issue.
- The expected and observed behavior.
- Any suggested mitigation or fix, if available.

Remove credentials, tokens, and other sensitive data from example inputs. You do
not need a complete exploit or a proposed fix to submit a report.

## Coordinated disclosure

Please keep vulnerability details private while maintainers assess the report
and coordinate a fix and disclosure with you. Updates and questions should stay
in the private report. Response and resolution times depend on maintainer
availability and the complexity of the issue.

Published security advisories are available on the repository's
[Security Advisories page](https://github.com/synackd/go-kargs/security/advisories).
