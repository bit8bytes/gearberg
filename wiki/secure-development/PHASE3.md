# PHASE 3

## IMP 1: Security APIs

Before going live, we will implement GitHub workflows to check periodically for security vulnerabilities.

## IMP 2: Secure handling of Sourcecode

Goals: Confidentiality, Integrity, and Availability.

- Access: Only [MAINTAINERS](./../../MAINTAINERS) have acccess to the repository. Codeowners will be added if multiple people have access to thiy repository.
- Rigorous Change Management: There is no process defined currently due to the early start of the project.
- Development Environment: Currently, the secure development environment is managed via Nix.
- Separability: Not applicable.
- Code Signatures: Only authenticated commits are accepeted.