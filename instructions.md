# General Purpose Event & Timeline Intelligence Platform

# Permanent Coding-Agent Instructions

## 1. Purpose of this file

This file contains permanent development instructions for the project.

A coding agent MUST read and follow this file before modifying or generating project code.

These instructions apply to the entire project unless a later, explicit user instruction overrides a particular requirement.

The project is expected to grow into a large software system. Do not optimize for rapid implementation at the expense of architecture, maintainability, security, testability, portability, or future extensibility.

---

# 2. Project vision

The software is a General Purpose Event & Timeline Intelligence Platform.

It is intended to allow users to import large quantities of primarily text-based chronological data, transform that data using user-defined rules, validate it, enrich it with metadata, correlate different streams of events, query it, and visualize the results.

Examples of possible future data include:

* symptom diaries
* food consumption
* nutrition
* purchases
* personal finance
* exercise
* medication
* environmental observations
* research data
* logs
* machine-generated events
* arbitrary chronological text

The application must NOT be architecturally designed as a medical application, finance application, or symptom diary application.

Those are merely possible domains.

The underlying abstraction is:

```
source data
    ↓
events
    ↓
metadata
    ↓
relationships
    ↓
queries
    ↓
analysis
    ↓
visualizations
```

The system must therefore remain domain-independent.

---

# 3. Development philosophy

The project will be developed incrementally.

Never implement future functionality merely because it has been mentioned in project documentation.

When implementing a requested feature:

1. Implement only the requested feature.
2. Create appropriate interfaces for future functionality where necessary.
3. Keep components loosely coupled.
4. Avoid speculative implementation.
5. Do not create unnecessary abstractions merely for the sake of abstraction.
6. Do not rewrite functioning code unnecessarily.
7. Preserve backward compatibility unless a deliberate breaking change is requested.
8. Add automated tests for new functionality.
9. Update documentation when architecture or behavior changes.

The project may eventually become very large.

Prefer an architecture that can remain understandable when the codebase contains hundreds of modules.

---

# 4. Technology strategy

The initial implementation should use:

Backend:

* Go
* Gin

Frontend:

* Svelte
* TypeScript
* Vite

Database:

* SQLite

Authentication:

* server-side sessions
* secure cookies
* Argon2id password hashing

The architecture must permit future replacement of SQLite with another database such as PostgreSQL without rewriting the application layer.

Likewise, authentication, AI providers, visualization engines, and other major subsystems should be replaceable through interfaces where appropriate.

---

# 5. Architecture

Prefer a clean layered architecture.

A reasonable conceptual structure is:

```
Frontend / Presentation
        ↓
HTTP/API layer
        ↓
Application services
        ↓
Domain logic
        ↓
Repositories / infrastructure
        ↓
Database
```

Do not put business logic into Svelte components.

Do not put business logic into HTTP handlers.

Do not put SQL directly into unrelated application code.

Keep responsibilities separated.

---

# 6. Dependency injection

Use dependency injection for major services.

Examples include:

* authentication service
* user service
* session service
* audit logging service
* repository implementations
* database access
* rule engine
* parser
* data-processing engine
* AI provider
* visualization provider

Do not unnecessarily introduce a large dependency-injection framework.

Simple constructor injection is preferred where practical.

---

# 7. Domain model

The eventual core domain should be capable of representing generalized events.

Do not hard-code assumptions such as:

```
symptom = medical event
```

or

```
expense = finance event
```

Instead, design generalized concepts such as:

* Dataset
* Source
* Record
* Event
* Timestamp
* Attribute
* Category
* Relationship
* Rule
* Transformation
* Validation
* Query
* View
* Visualization

The exact model may evolve.

Do not prematurely implement the entire domain model.

---

# 8. Raw-data preservation

Imported source data must eventually be treated as immutable source material.

Never destroy the original source representation merely because a transformation has been applied.

Transformations should be represented in a manner that permits:

* traceability
* auditing
* reproducibility
* undo
* reprocessing
* comparison of versions

The eventual system must allow a user to trace processed information back to its original source record.

---

# 9. Large-data requirement

The application may eventually process millions of lines.

Performance is therefore a first-class requirement.

Avoid:

* loading entire datasets into browser memory
* sending millions of records to the frontend
* unnecessary database queries
* N+1 queries
* repeated parsing of identical data
* unnecessary JSON serialization
* unnecessary copying of large strings
* expensive operations inside UI rendering

The frontend should request only the information necessary to display the current view.

Use pagination, cursor-based retrieval, streaming, indexing, caching, virtualization, or other appropriate mechanisms.

---

# 10. Frontend principles

The frontend is primarily a presentation layer.

It should handle:

* rendering
* user interaction
* navigation
* input collection
* selection
* drag/drop
* scrolling
* responsive layout

The backend should handle:

* parsing
* validation
* transformations
* calculations
* searching
* indexing
* correlation
* rule execution
* database operations
* security-sensitive operations

Do not move computationally expensive business logic into the browser merely because it is convenient.

---

# 11. Virtualized display

Large datasets must never be rendered entirely at once.

When displaying large lists or tables, use virtualization such as TanStack Virtual or an equivalent mature solution.

Only rows that can reasonably be displayed in the current viewport should be rendered.

The architecture should eventually support:

* vertical scrolling
* horizontal scrolling
* virtualized rows
* large horizontally wide records
* minimaps
* search results
* error markers
* selections

---

# 12. Responsive design

The UI must work on:

* desktop
* laptop
* tablet
* smartphone

It must respond correctly to:

* different viewport widths
* different viewport heights
* portrait orientation
* landscape orientation
* touch input
* mouse input where appropriate

The default visual design should be:

* dark
* minimalist
* modern
* vertically compact
* space efficient

Blue should be the primary accent.

Do not use colors merely for decoration.

Use color primarily for:

* warnings
* errors
* selections
* categories
* important states

---

# 13. Offline/self-contained operation

The application must eventually be capable of running locally without Internet access after installation.

Do not make runtime requests to:

* Google Fonts
* CDN-hosted JavaScript
* CDN-hosted CSS
* remote icon libraries
* remote analytics
* unnecessary external APIs

Frontend dependencies required by the application must be bundled with the application.

AI/API functionality must be optional.

The core application must work without Internet access.

---

# 14. Authentication

Use server-side sessions rather than JWT for normal authentication.

Session identifiers must be:

* cryptographically random
* sufficiently long
* stored securely
* associated with server-side session records
* revocable
* expired after an appropriate period
* rotated when appropriate

Use secure cookies with appropriate:

* HttpOnly
* SameSite
* Secure

attributes.

Do not expose session identifiers through URLs.

Do not store authentication tokens in localStorage unless there is a compelling, documented reason.

---

# 15. Passwords

Never store plaintext passwords.

Use Argon2id with:

* cryptographically secure random salts
* appropriate parameters
* server-side verification

Password requirements should follow current security best practices.

Encourage passphrases.

Do not impose unnecessarily restrictive rules such as requiring arbitrary combinations of symbols when a long passphrase provides better security.

Reject obviously weak passwords such as:

* common passwords
* very short passwords
* repeated characters
* obvious sequences
* passwords containing the user's email/name in an obvious way

A password confirmation field is not mandatory.

A password visibility toggle may be provided.

---

# 16. Authentication attack resistance

Implement appropriate defenses against:

* brute-force login attempts
* credential stuffing
* account enumeration
* automated bots
* session fixation
* CSRF
* XSS
* SQL injection
* request flooding

Failed login attempts should introduce an intentional delay.

Repeated failures should result in progressively stronger throttling/rate limiting.

Do not reveal whether an email address belongs to an account.

For example, avoid messages such as:

```
Email does not exist.
```

Use a generic authentication failure message.

---

# 17. Logging and auditing

Security-relevant events must be logged.

Eventually the database should contain structured logs/audit records including appropriate information such as:

* timestamp
* event type
* user ID where known
* IP address where appropriate
* success/failure
* relevant non-sensitive metadata

Never log:

* passwords
* password hashes
* session identifiers
* authentication secrets
* sensitive data unnecessarily

Logs should eventually support rotation/retention.

Administrative logs must not be visible to ordinary users.

---

# 18. Privacy

The application may eventually store highly sensitive information.

Therefore:

* minimize collected information
* minimize transmitted information
* minimize logged information
* encrypt sensitive data where appropriate
* use secure transport for network deployments
* provide appropriate account/data controls
* avoid sending user data to external services unless explicitly requested

Do not claim legal compliance automatically.

Where legal/regulatory requirements are relevant, identify them as requirements to be evaluated rather than making unsupported compliance claims.

---

# 19. Database

SQLite is the initial database.

Use:

* appropriate indexes
* transactions
* foreign keys
* constraints
* prepared/parameterized statements
* migrations

Do not assume SQLite must remain the permanent database.

Database access must be isolated behind repository/data-access boundaries so that PostgreSQL or another database can be introduced later.

---

# 20. API design

The API should be designed with security and future scalability in mind.

Prefer state-changing operations through POST/PUT/PATCH/DELETE rather than GET.

GET may be used where semantically appropriate for safe retrieval.

Do not attempt to avoid GET merely for security reasons; security must instead come from:

* TLS
* authentication
* authorization
* secure cookies
* CSRF protection
* proper API design

Never transmit sensitive information in URLs unnecessarily.

---

# 21. Error handling

Errors must be handled explicitly.

Do not allow unexpected errors to crash the application.

Users should receive useful but non-sensitive error messages.

Detailed errors should be available to authorized administrators through logs.

Never expose:

* stack traces
* SQL statements containing sensitive information
* filesystem paths unnecessarily
* internal implementation details

---

# 22. Testing

Every significant feature must have automated tests.

Tests should include, where appropriate:

* unit tests
* integration tests
* database tests
* API tests
* authentication tests
* security tests
* frontend tests

When a new rule type is eventually implemented, create representative positive and negative test cases.

Do not consider a feature complete merely because it works manually.

---

# 23. Build system

The project must provide a Python 3 development/build orchestration script.

The programmer should ideally be able to run:

```
python3 build.py
```

and obtain a menu providing common development operations.

The script should eventually support:

* OS detection
* dependency checking
* dependency installation
* build configuration
* debug/release selection
* frontend build
* backend build
* complete build
* cleaning
* tests
* database initialization
* application startup
* health checks
* diagnostics

The script must return to the menu after completing each operation.

Only Exit should terminate the menu.

Do not hide failures.

Clearly display command failures and their causes.

---

# 24. Operating-system support

The application should eventually work locally on:

* Linux
* Windows
* macOS

Do not assume:

* bash exists
* apt exists
* systemd exists
* a particular filesystem layout
* Unix-specific commands

The Python orchestration layer should detect the OS and use appropriate commands.

The application itself should avoid unnecessary platform-specific assumptions.

---

# 25. Source-code organization

Keep source files reasonably small.

Do not create enormous files containing unrelated functionality.

Use meaningful directory structures.

Prefer cohesive modules.

Avoid circular dependencies.

Do not duplicate functionality when it can reasonably be shared.

---

# 26. Future rule engine

The eventual application will contain a powerful rule engine.

Future rules may perform operations such as:

* recognizing dates
* recognizing times
* extracting timestamps
* normalizing terminology
* splitting phrases
* identifying categories
* recognizing numerical values
* applying substitutions
* examining surrounding records
* examining time windows
* identifying patterns
* validating chronological order
* correlating events
* performing calculations

Do not implement this engine until explicitly requested.

When it is implemented, keep it independent from the UI.

---

# 27. Future AI integration

AI is an optional processing mechanism.

The architecture should eventually support:

* local LLMs
* API-based LLMs
* multiple AI providers
* AI-assisted rule creation
* AI-assisted queries
* AI-assisted classification

AI must not be a mandatory dependency for the core application.

Where AI proposes modifications to user data, the architecture should support review/approval before destructive changes.

---

# 28. Future visualization

The eventual application may support:

* 2D charts
* 3D visualization
* timelines
* graphs
* networks
* maps
* mini-maps
* correlations

Possible technologies include:

* WebGL
* Three.js
* SVG
* Canvas

Do not introduce a visualization framework until required.

---

# 29. Future collaboration

The application is initially intended primarily for a single user.

Future versions may support multiple users simultaneously.

Possible future functionality includes:

* authorization by the primary user
* simultaneous viewing
* shared cursors
* cursor colors
* text chat
* voice
* video
* collaborative annotations

Do not implement collaboration now.

However, avoid architectural decisions that make multi-user access impossible.

---

# 30. Account deactivation

Unregistering an account does not initially mean immediate physical deletion.

The system should support an account-deactivation state.

Future retention/deletion functionality may permanently remove data according to applicable legal and contractual requirements.

Do not hard-code a particular legal retention period unless explicitly specified by the application owner after legal review.

---

# 31. Administrator access

Future versions may support an administrator accessing a user's account without knowing the user's password, but only after explicit authorization and with complete auditing and notification.

This functionality must:

* never expose the user's password
* never impersonate silently
* notify the affected user
* create an audit record
* clearly identify the administrator access

Do not implement this unless explicitly requested.

---

# 32. Code-generation discipline

When asked to implement a feature:

1. Inspect the existing architecture.
2. Read relevant documentation.
3. Identify existing interfaces/services that should be reused.
4. Do not duplicate existing functionality.
5. Implement the smallest coherent change.
6. Add tests.
7. Run the relevant tests.
8. Run formatting/linting where available.
9. Report failures honestly.
10. Update documentation if necessary.

Never pretend that tests passed if they were not actually run.

Never claim that an external dependency was installed if installation failed.

---

# 33. No unnecessary technology

Do not add a library merely because it is popular.

Before introducing a dependency, consider:

* Is it actually necessary?
* Is it maintained?
* Is its license appropriate?
* Does it increase security risk?
* Does it increase build complexity?
* Does it work offline after installation?
* Is the functionality easy to implement safely without it?
* Will it make future maintenance harder?

Prefer mature, well-maintained dependencies.

---

# 34. Documentation

Maintain documentation explaining important architectural decisions.

When an architectural decision is non-obvious, document:

* why it was chosen
* alternatives considered
* consequences

Do not create documentation that merely repeats the source code.

---

# 35. Important scope rule

The existence of a future requirement does NOT authorize implementing it now.

For example, if the instructions mention:

* AI
* visualization
* nutrition
* finance
* medical data
* collaboration
* WebGL
* plugins

that does not mean those features should be implemented during an unrelated task.

Implement only what the current user request asks for.

---

# 36. Priority order

When requirements conflict, prioritize:

1. Security
2. Data integrity
3. Correctness
4. Privacy
5. Maintainability
6. Testability
7. Performance
8. Usability
9. Feature breadth

Do not sacrifice security or data integrity merely to make a feature easier to implement.

---

# 37. Agent communication

Before making a major architectural change, explain the proposed change and why it is necessary.

For ordinary implementation work, do not unnecessarily ask the user for permission for every small decision.

If a requirement is genuinely ambiguous and choosing incorrectly could cause substantial rework, ask for clarification.

Otherwise make a reasonable conservative choice and document it.

---

# 38. Current implementation scope

The current implementation stage is intentionally small.

The initial application consists of:

* project setup
* development/build automation
* SQLite database
* user registration
* login
* logout
* account deactivation
* basic authenticated landing page
* basic navigation
* settings placeholder
* help placeholder
* FAQ placeholder
* logs placeholder
* about placeholder

Do not implement the future event-processing platform until explicitly requested.

