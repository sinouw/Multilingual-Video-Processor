# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by:
- Opening a private security advisory on GitHub: https://github.com/sinouw/multilingual-video-processor/security/advisories
- Or by contacting the maintainer: [@sinouw](https://github.com/sinouw)

**Please do not create a public GitHub issue for security vulnerabilities.**

Include the following information:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will respond within 48 hours and work with you to resolve the issue.

## Security Best Practices

When deploying this service:

1. **API Keys**: Store API keys securely, never commit them to version control
2. **Service Accounts**: Use service accounts with least privilege principle
3. **CORS**: Configure CORS origins appropriately for your use case
4. **Rate Limiting**: Enable rate limiting to prevent abuse
5. **Monitoring**: Monitor logs for suspicious activity
6. **Updates**: Keep dependencies up to date

## Known Security Considerations

- The service processes user-provided video files - validate inputs carefully
- Temporary files are created during processing - ensure proper cleanup
- API keys are required - use secure secret management
- Service runs with service account credentials - follow principle of least privilege