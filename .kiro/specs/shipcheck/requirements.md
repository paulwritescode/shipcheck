# Requirements Document

## Introduction

ShipCheck is an evidence-based launch-readiness workspace for small product and engineering teams. It consolidates a launch brief, checklist, risk register, evidence, and an explainable readiness assessment into a single workspace so a team can answer one question before releasing: are we ready to ship, and what still needs attention?

The core object is a **Launch**. Each launch aggregates requirements, checklist items, risks, evidence links, and launch notes, and from those inputs ShipCheck computes a **readiness status** (Ready, Conditionally Ready, Not Ready, or Needs Review) together with an explanation of why that status was assigned. The workspace supports creating launches, seeding a demonstration launch, editing checklist items and risks, running an AI analysis that produces recommendations, and publishing a public read-only readiness report.

A core principle governs the relationship between the deterministic logic and the AI: the deterministic Readiness_Engine decides the readiness state, and the AI agent (the ShipCheck_Advisor) explains the state and recommends what to do next. The AI never approves a release, never deletes data, never changes statuses, and never independently contacts external services.

This document specifies the requirements for the first release. The readiness scoring and status-determination rules are defined with special precision because they are the core logic of the product and will drive property-based tests. The AI advisor, model-provider abstraction, MCP external context, and the security, privacy, and compliance boundaries are specified so that the advisor stays read-only and credentials and untrusted content are handled safely. All requirements follow EARS patterns and INCOSE quality rules.

## Glossary

- **ShipCheck**: The overall system described by this document.
- **Launch**: The primary domain object representing a product release, feature release, beta, campaign, migration, or delivery milestone. A Launch owns a name, description, target date, owner, checklist items, risks, evidence, launch notes, and a computed readiness assessment.
- **Launch_Owner**: The person accountable for a Launch. Stored as a named value on the Launch.
- **Checklist_Item**: A unit of launch work with a category, owner, priority, completion state, optional evidence, and a critical flag.
- **Completion_State**: The state of a Checklist_Item, one of complete, in progress, blocked, or incomplete. For readiness scoring, in progress, blocked, and incomplete all count as "not complete"; only complete counts as complete.
- **Risk**: A register entry with a title, description, severity, likelihood, owner, mitigation, status, and optional due date.
- **Evidence**: A stored link or note that supports a Checklist_Item or Risk. Includes a URL or free-text note.
- **Category**: One of the seven fixed groupings: Requirements, Engineering, Quality Assurance, Documentation, Ownership, Risks, Launch Operations.
- **Priority**: A Checklist_Item ranking, one of Low, Medium, High.
- **Severity**: A Risk impact level, one of Low, Medium, High.
- **Likelihood**: A Risk probability level, one of Low, Medium, High.
- **Critical_Item**: A Checklist_Item whose critical flag is true.
- **Blocking_Risk**: A Risk with Severity of High and a status that is not Resolved and not Accepted.
- **Readiness_Status**: The computed launch state, one of Ready, Conditionally Ready, Not Ready, Needs Review.
- **Readiness_Assessment**: The Readiness_Status plus the ordered list of reasons that produced it and the per-Category completion states.
- **Readiness_Engine**: The component that computes the Readiness_Assessment from Launch data.
- **AI_Analyzer**: The component that reviews Launch data and returns recommendations. Realized by the ShipCheck_Advisor.
- **ShipCheck_Advisor**: The read-only advisor agent that explains the computed readiness state and recommends next actions. The ShipCheck_Advisor never approves or denies a release, never changes a Readiness_Status, never deletes data, never mutates Launch data, and never independently contacts external services.
- **ModelProvider**: The provider-agnostic interface through which ShipCheck accesses an AI model. Concrete providers include a deterministic fallback provider (the always-free default the live demo runs on, using no external model and no key), a Hugging Face SLM provider (opt-in for the demo), an Amazon Bedrock provider (alternative opt-in), an NVIDIA catalog provider (alternative opt-in), and a local model provider, all selectable without changing product code.
- **HuggingFace_Provider**: The opt-in ModelProvider implementation that calls a small hosted small language model (SLM), such as Llama 3.2 1B or Qwen2.5 1.5B, via the Hugging Face Inference API using a server-side API key. The Hugging Face free serverless allowance is rate-limited and intended for light, non-production use.
- **Bedrock_Provider**: The opt-in ModelProvider implementation that calls Amazon Bedrock (e.g., the Amazon Nova Micro model) from the server-side Go Lambda using the Lambda's IAM role for authentication (no API key). Bedrock has no free-inference tier; it is free only while AWS signup or Activate credits remain, then it bills per token.
- **MCP_Tool**: A read-only tool exposed over the Model Context Protocol that the application invokes to fetch external context for the ShipCheck_Advisor.
- **Untrusted_External_Content**: Any content fetched from an external source, including repository files, issues, pull requests, documentation pages, and web pages. Untrusted_External_Content is treated as data only and never as instructions.
- **Launch_Snapshot**: The immutable copy of Launch data, including the Readiness_Assessment computed by the Readiness_Engine, that is passed to the ShipCheck_Advisor for a single analysis request.
- **Public_Report**: The read-only, externally shareable view of a Launch.
- **Share_Link**: The unique URL that grants access to a Public_Report.
- **Seeded_Launch**: The pre-populated demonstration Launch named "Team Inbox 2.0".
- **Visitor**: A person viewing ShipCheck without an authenticated account.

## Requirements

### Requirement 1: Create a Launch

**User Story:** As a small-team product lead, I want to create a launch with basic details, so that I have a workspace to track release readiness.

#### Acceptance Criteria

1. WHEN a user submits the launch creation form with a launch name of 1 to 200 characters, a description of 0 to 5,000 characters, a target date that is a valid calendar date, and a Launch_Owner, THE ShipCheck SHALL create a new Launch and persist it to storage.
2. IF a user submits the launch creation form with a launch name that is empty or contains only whitespace, THEN THE ShipCheck SHALL reject the submission and display a validation message identifying the launch name as required.
3. IF a user submits the launch creation form with a launch name longer than 200 characters, THEN THE ShipCheck SHALL reject the submission and display a validation message identifying the launch name as exceeding the maximum length.
4. IF a user submits the launch creation form with an empty target date, THEN THE ShipCheck SHALL reject the submission and display a validation message identifying the target date as required.
5. IF a user submits the launch creation form with a target date that is not a valid calendar date, THEN THE ShipCheck SHALL reject the submission and display a validation message identifying the target date as invalid.
6. WHERE the user provides an optional product area, THE ShipCheck SHALL store the product area on the Launch.
7. WHERE the user provides an optional repository URL or documentation URL, THE ShipCheck SHALL store the URL on the Launch.
8. IF a user provides a repository URL or documentation URL that is not a syntactically valid URL, THEN THE ShipCheck SHALL reject the submission and display a validation message identifying the URL as invalid.
9. WHEN a Launch is created, THE ShipCheck SHALL assign the Launch a Readiness_Status of Needs Review until a Readiness_Assessment is computed.
10. WHEN a user chooses to start from the Seeded_Launch, THE ShipCheck SHALL create a Launch pre-populated with the demonstration data defined in Requirement 12.

### Requirement 2: Maintain the Launch Brief

**User Story:** As a launch owner, I want to record the launch context, so that the readiness review reflects what is being released and why.

#### Acceptance Criteria

1. WHEN a user saves launch brief content, THE ShipCheck SHALL store the brief on the Launch and persist it to storage.
2. THE ShipCheck SHALL allow the launch brief to contain what is being released, the target audience, the definition of success, requirements, known constraints, and dependencies.
3. WHEN a user edits and saves an existing launch brief, THE ShipCheck SHALL replace the stored brief with the edited content.
4. WHILE a launch brief is empty, THE ShipCheck SHALL display an empty-state prompt describing how to add brief content.

### Requirement 3: Manage Checklist Items

**User Story:** As a team member preparing a release, I want to manage categorized checklist items with owners and evidence, so that the remaining work is explicit and assignable.

#### Acceptance Criteria

1. WHEN a user creates a Checklist_Item with a title and a Category, THE ShipCheck SHALL add the Checklist_Item to the Launch and persist it.
2. THE ShipCheck SHALL constrain each Checklist_Item Category to exactly one of the seven values: Requirements, Engineering, Quality Assurance, Documentation, Ownership, Launch Operations, Risks.
3. WHEN a user assigns an owner to a Checklist_Item, THE ShipCheck SHALL store the owner on the Checklist_Item.
4. WHEN a user sets a Priority on a Checklist_Item, THE ShipCheck SHALL store the Priority as one of Low, Medium, or High.
5. WHEN a user marks a Checklist_Item complete, THE ShipCheck SHALL set the completion state of the Checklist_Item to complete and record the change.
6. WHEN a user marks a completed Checklist_Item incomplete, THE ShipCheck SHALL set the completion state of the Checklist_Item to incomplete.
7. WHEN a user sets the critical flag on a Checklist_Item, THE ShipCheck SHALL designate the Checklist_Item as a Critical_Item.
8. WHEN a user attaches an Evidence link or note to a Checklist_Item, THE ShipCheck SHALL store the Evidence on the Checklist_Item.
9. WHEN a user edits a Checklist_Item field, THE ShipCheck SHALL persist the edited value and preserve all other fields of the Checklist_Item.
10. WHEN a user deletes a Checklist_Item, THE ShipCheck SHALL remove the Checklist_Item from the Launch.
11. WHEN a user applies an incomplete-only filter, THE ShipCheck SHALL display only Checklist_Items whose completion state is incomplete.
12. WHEN a user applies a high-priority filter, THE ShipCheck SHALL display only Checklist_Items whose Priority is High.
13. THE ShipCheck SHALL group Checklist_Items by Category in the checklist view.

### Requirement 4: Manage the Risk Register

**User Story:** As a launch owner, I want a risk register with severity, likelihood, and ownership, so that dangers to the release are visible and assigned.

#### Acceptance Criteria

1. WHEN a user creates a Risk with a title, THE ShipCheck SHALL add the Risk to the Launch and persist it.
2. WHEN a user sets a Severity on a Risk, THE ShipCheck SHALL store the Severity as one of Low, Medium, or High.
3. WHEN a user sets a Likelihood on a Risk, THE ShipCheck SHALL store the Likelihood as one of Low, Medium, or High.
4. THE ShipCheck SHALL allow each Risk to store a description, an owner, a mitigation, a status, and an optional due date.
5. THE ShipCheck SHALL constrain each Risk status to one of Open, Mitigating, Resolved, or Accepted.
6. WHERE a Risk has a Severity of High or a Likelihood of High, THE ShipCheck SHALL apply a distinct high-emphasis visual treatment to that Risk in the risk register view.
7. WHEN a user edits a Risk field, THE ShipCheck SHALL persist the edited value and preserve all other fields of the Risk.
8. WHEN a user deletes a Risk, THE ShipCheck SHALL remove the Risk from the Launch.

### Requirement 5: Attach Evidence and Links

**User Story:** As a reviewer, I want evidence attached to checklist items and risks, so that completion and mitigation claims are supported by references.

#### Acceptance Criteria

1. WHEN a user adds an Evidence entry with a URL, THE ShipCheck SHALL store the URL and associate the Evidence with the target Checklist_Item or Risk.
2. WHEN a user adds an Evidence entry with a note, THE ShipCheck SHALL store the note text and associate the Evidence with the target Checklist_Item or Risk.
3. IF a user submits an Evidence URL that is not a syntactically valid URL, THEN THE ShipCheck SHALL reject the entry and display a validation message identifying the URL as invalid.
4. THE ShipCheck SHALL support Evidence references to GitHub pull requests, GitHub issues, documentation pages, design files, test reports, deployment URLs, and decision records by storing their links and notes.
5. WHEN a user removes an Evidence entry, THE ShipCheck SHALL delete the Evidence from its associated Checklist_Item or Risk.

### Requirement 6: Compute Readiness Status

**User Story:** As a launch owner, I want an automatically computed readiness status, so that I know whether the launch is Ready, Conditionally Ready, Not Ready, or Needs Review.

#### Acceptance Criteria

1. WHEN a user creates, edits, or deletes a Checklist_Item, a Risk, or an Evidence entry on the Launch, or when a user sets or clears the Launch_Owner, THE Readiness_Engine SHALL recompute the Readiness_Assessment for the Launch.
2. THE Readiness_Engine SHALL treat a Checklist_Item Completion_State of complete as complete for scoring, and SHALL treat a Completion_State of in progress, blocked, or incomplete as not complete for scoring.
3. THE Readiness_Engine SHALL evaluate the readiness rules in the following fixed precedence order and assign the first matching Readiness_Status: Needs Review, then Not Ready, then Conditionally Ready, then Ready.
4. IF the Launch has zero Checklist_Items and zero Risks, THEN THE Readiness_Engine SHALL assign the Readiness_Status Needs Review.
5. IF the Launch has no Launch_Owner, THEN THE Readiness_Engine SHALL assign the Readiness_Status Needs Review.
6. IF a Critical_Item has a Completion_State of complete AND that Critical_Item has zero associated Evidence entries, THEN THE Readiness_Engine SHALL assign the Readiness_Status Needs Review, because a completion claim without evidence is contradictory.
7. IF at least one Critical_Item is not complete for scoring, THEN THE Readiness_Engine SHALL assign the Readiness_Status Not Ready.
8. IF at least one Blocking_Risk exists, THEN THE Readiness_Engine SHALL assign the Readiness_Status Not Ready.
9. IF at least one Checklist_Item in the Launch Operations Category is a Critical_Item AND that Checklist_Item is not complete for scoring, THEN THE Readiness_Engine SHALL assign the Readiness_Status Not Ready, covering the missing rollback or contingency case.
10. IF every Critical_Item is complete for scoring with at least one Evidence entry AND no Blocking_Risk exists AND at least one non-critical Checklist_Item is not complete for scoring, THEN THE Readiness_Engine SHALL assign the Readiness_Status Conditionally Ready.
11. IF every Critical_Item is complete for scoring with at least one Evidence entry AND no Blocking_Risk exists AND at least one Risk has a status of Accepted, THEN THE Readiness_Engine SHALL assign the Readiness_Status Conditionally Ready.
12. IF every Checklist_Item is complete for scoring AND every Critical_Item has at least one Evidence entry AND no Blocking_Risk exists AND the Launch has a Launch_Owner, THEN THE Readiness_Engine SHALL assign the Readiness_Status Ready.
13. WHERE no rule of higher precedence in this requirement matches the current Launch state, THE Readiness_Engine SHALL assign the Readiness_Status Ready as the default, so that every possible Launch state receives exactly one Readiness_Status.
14. THE Readiness_Engine SHALL assign exactly one Readiness_Status to a Launch for a given Launch state.
15. WHEN the Readiness_Engine evaluates two Launch states that are identical, THE Readiness_Engine SHALL assign an identical Readiness_Status and an identical set of reason entries to both.

### Requirement 7: Explain the Readiness Assessment

**User Story:** As a launch owner and as a reviewer, I want to see why a readiness status was assigned, so that the decision is transparent and actionable.

#### Acceptance Criteria

1. WHEN the Readiness_Engine assigns a Readiness_Status, THE Readiness_Engine SHALL produce at least one reason entry that identifies the matched readiness rule from Requirement 6 that determined the assigned Readiness_Status.
2. WHERE a reason entry was triggered by a condition evaluated on a specific Checklist_Item or Risk, THE Readiness_Assessment SHALL include the identifier of that Checklist_Item or Risk on the reason entry, and WHERE a reason entry was triggered by a launch-level condition not tied to a single item, THE Readiness_Assessment SHALL omit an item identifier from that reason entry.
3. WHEN the Readiness_Status is Not Ready, THE Readiness_Engine SHALL include exactly one reason entry for each incomplete Critical_Item and exactly one reason entry for each Blocking_Risk.
4. WHEN the Readiness_Status is Needs Review, THE Readiness_Engine SHALL include one reason entry for each condition that triggered the status, covering an absent Launch_Owner, a Launch with zero Checklist_Items and zero Risks, and each Critical_Item that is complete with zero associated Evidence entries.
5. THE Readiness_Engine SHALL provide, for each reason entry that reports an incomplete Critical_Item, a Blocking_Risk, an absent Launch_Owner, or a completion claim with zero Evidence entries, a recommended next action that names the specific Checklist_Item, Risk, or missing attribute the user must change to resolve that reason entry.
6. THE Readiness_Engine SHALL compute, for each of the seven Categories, a completion state consisting of the count of complete Checklist_Items in that Category and the total count of Checklist_Items in that Category, and SHALL classify the Category as complete when all its Checklist_Items are complete, partial when at least one but not all are complete, and empty when the Category contains zero Checklist_Items.

### Requirement 8: Readiness Dashboard

**User Story:** As a launch owner, I want a dashboard summarizing the launch, so that I can see status, progress, blockers, risks, and next actions at a glance.

#### Acceptance Criteria

1. THE ShipCheck SHALL display on the dashboard the launch title, the target date, the overall Readiness_Status, and the overall completion progress.
2. THE ShipCheck SHALL display on the dashboard a per-Category completion breakdown for the seven Categories.
3. THE ShipCheck SHALL display on the dashboard the current critical blockers, defined as incomplete Critical_Items and Blocking_Risks.
4. THE ShipCheck SHALL display on the dashboard the Risks ordered so that Risks with High Severity appear before Risks with lower Severity.
5. THE ShipCheck SHALL display on the dashboard the recommended next actions produced by the Readiness_Engine.
6. WHEN a user opens a critical blocker from the dashboard, THE ShipCheck SHALL display the reason the blocker matters and its recommended next action.
7. WHILE a Launch has zero Checklist_Items and zero Risks, THE ShipCheck SHALL display an empty-state prompt guiding the user to add launch data.

### Requirement 9: AI Launch Analysis via the ShipCheck Advisor

**User Story:** As a launch owner, I want a read-only AI advisor that explains the computed readiness state and recommends next actions, so that I receive guidance about gaps and blockers while I keep control of all changes and the deterministic engine remains the sole authority for the status.

#### Acceptance Criteria

1. WHEN a user runs the launch analysis, THE ShipCheck_Advisor SHALL review the Launch data and return recommendations covering missing information, likely blockers, possible contradictions, unowned work, suggested Checklist_Items, and suggested launch-review questions.
2. THE ShipCheck_Advisor SHALL present its output as recommendations that require explicit user action to apply.
3. WHEN a user accepts a suggested Checklist_Item, THE ShipCheck SHALL create the corresponding Checklist_Item on the Launch.
4. WHEN a user dismisses a recommendation, THE ShipCheck SHALL discard that recommendation without modifying the Launch.
5. IF the launch analysis cannot be completed, THEN THE ShipCheck SHALL display an error state describing that the analysis did not complete and SHALL leave the Launch data unchanged.
6. WHILE the launch analysis is running, THE ShipCheck SHALL display a loading state.
7. THE ShipCheck_Advisor SHALL operate on a Launch_Snapshot produced after the Readiness_Engine computes the Readiness_Assessment, and THE Readiness_Engine SHALL remain the sole authority for the Readiness_Status.
8. THE ShipCheck_Advisor SHALL NOT approve or deny a release, SHALL NOT change any Readiness_Status, SHALL NOT delete data, and SHALL NOT mutate Launch data.
9. THE ShipCheck SHALL provide the following focused advisor actions rather than a generic chatbot: Analyze Launch, Find Blockers, Suggest Checklist, Review Risks, Prepare Launch Review, Explain Readiness, and Re-analyze After Changes.
10. WHEN the ShipCheck_Advisor returns a result, THE ShipCheck_Advisor SHALL return structured JSON containing a list of recommendations, where each recommendation includes a type, a priority, a title, a reason, a suggested_action, and evidence references to real Launch entities, together with a list of suggested_questions.
11. WHEN the ShipCheck backend receives an advisor response, THE ShipCheck SHALL validate the response against a defined schema before displaying the response.
12. IF an advisor recommendation references a Launch entity that does not exist on the Launch, THEN THE ShipCheck SHALL discard that recommendation and record an error for that recommendation.
13. WHEN a user runs the Prepare Launch Review action, THE ShipCheck_Advisor SHALL produce a concise review brief containing the current Readiness_Status, the completed work, the unresolved blockers, the accepted Risks, the decisions needed, and the questions for stakeholders.
14. WHEN a user runs the Re-analyze After Changes action, THE ShipCheck_Advisor SHALL report the Launch changes that occurred since the previous analysis using the recent-launch-changes capability.

### Requirement 10: Public Read-Only Share Report

**User Story:** As a launch owner, I want a public read-only report with a unique link, so that stakeholders outside the workspace can understand launch readiness without exposing private details.

#### Acceptance Criteria

1. WHEN a user enables sharing for a Launch, THE ShipCheck SHALL generate a unique Share_Link that resolves to the Public_Report for that Launch.
2. WHEN a Visitor opens a valid Share_Link, THE ShipCheck SHALL display the Public_Report without requiring authentication.
3. THE Public_Report SHALL display the launch name, the target date, the overall Readiness_Status, the per-Category scores, the completed and incomplete work, the high-priority Risks, the open questions, and the last-updated timestamp.
4. THE Public_Report SHALL exclude private launch notes and evidence marked private from the displayed content.
5. WHEN a user regenerates the Share_Link, THE ShipCheck SHALL issue a new Share_Link and SHALL make the previous Share_Link no longer resolve to the Public_Report.
6. WHEN a user disables sharing, THE ShipCheck SHALL make the Share_Link no longer resolve to the Public_Report.
7. IF a Visitor opens a disabled or regenerated Share_Link, THEN THE ShipCheck SHALL display a message indicating the report is unavailable.
8. THE Public_Report SHALL present the same Readiness_Status that the Readiness_Engine computed for the underlying Launch.

### Requirement 11: Landing Page and Guest Exploration

**User Story:** As a first-time visitor, I want a landing page and the ability to explore the demo without an account, so that I can understand the product within one minute.

#### Acceptance Criteria

1. THE ShipCheck SHALL provide a landing page that describes the product purpose and provides an entry point to the Seeded_Launch.
2. WHEN a Visitor chooses to explore the demonstration, THE ShipCheck SHALL open the Seeded_Launch without requiring the Visitor to create an account.
3. WHILE a Visitor explores the Seeded_Launch without an account, THE ShipCheck SHALL allow read access to the launch dashboard, checklist, risk register, and readiness assessment.

### Requirement 12: Seeded Demonstration Launch

**User Story:** As a demonstrator, I want a realistic pre-populated launch, so that the product story is understandable in under one minute.

#### Acceptance Criteria

1. THE ShipCheck SHALL provide a Seeded_Launch named "Team Inbox 2.0" with a target date of October 16, 2026.
2. THE Seeded_Launch SHALL contain 18 Checklist_Items, of which 13 have completion state complete, 3 are in progress, and 2 are blocked.
3. THE Seeded_Launch SHALL contain 4 Risks, of which exactly one has Severity High.
4. THE Seeded_Launch SHALL contain at least one Checklist_Item with no assigned owner.
5. THE Seeded_Launch SHALL contain an incomplete rollback documentation Checklist_Item in the Launch Operations Category.
6. THE Seeded_Launch SHALL include a public preview URL and Evidence entries referencing example GitHub issues and pull requests.
7. WHEN the Readiness_Engine evaluates the Seeded_Launch in its initial state, THE Readiness_Engine SHALL assign the Readiness_Status Not Ready.
8. WHEN a user assigns an owner and Evidence to the top blocker of the Seeded_Launch, marks the related Critical_Item complete, resolves the Blocking_Risk, and re-runs readiness, THE Readiness_Engine SHALL assign the Readiness_Status Conditionally Ready.

### Requirement 13: Persistence

**User Story:** As a user, I want my launch data to persist, so that my work is retained across sessions.

#### Acceptance Criteria

1. WHEN a user creates or edits a Launch, a Checklist_Item, a Risk, or Evidence, THE ShipCheck SHALL persist the change to storage.
2. WHEN a user reopens a previously saved Launch, THE ShipCheck SHALL load the Launch with all previously persisted Checklist_Items, Risks, Evidence, and launch brief content.
3. IF a persistence operation fails, THEN THE ShipCheck SHALL display an error state describing that the change was not saved.

### Requirement 14: Responsive Layout and UI States

**User Story:** As a user on any device, I want a responsive layout with clear UI states, so that I can use ShipCheck on desktop and mobile with understandable feedback.

#### Acceptance Criteria

1. THE ShipCheck SHALL render a usable layout on both desktop viewport widths and mobile viewport widths.
2. WHILE data for a view is loading, THE ShipCheck SHALL display a loading state.
3. WHERE a view has no data to display, THE ShipCheck SHALL display an empty state describing how to add data.
4. WHERE a filter or search returns zero matching items, THE ShipCheck SHALL display a no-results state.
5. IF an operation returns an error, THEN THE ShipCheck SHALL display an error state describing the failure.

### Requirement 15: Scope Boundaries

**User Story:** As a product stakeholder, I want the first release scoped tightly, so that ShipCheck focuses on launch review rather than replacing existing tools.

#### Acceptance Criteria

1. THE ShipCheck SHALL store external references as links and notes rather than synchronizing state from external systems.
2. THE ShipCheck SHALL present the Readiness_Assessment as decision support, and SHALL require a human user to make any change to Launch data.
3. THE ShipCheck SHALL limit its first-release feature set to launch creation, launch brief, checklist management, risk tracking, evidence links, readiness scoring, AI analysis recommendations, public read-only reports, persistence, and responsive UI states.

### Requirement 16: Model Provider Abstraction and Key Safety

**User Story:** As a system operator, I want the AI model accessed through a provider-agnostic interface with server-side keys, so that ShipCheck can switch model providers without code changes and never exposes credentials to clients.

#### Acceptance Criteria

1. THE ShipCheck SHALL access the AI model through the provider-agnostic ModelProvider interface, so that a concrete provider such as a deterministic fallback provider, a Hugging Face SLM provider, an Amazon Bedrock provider, an NVIDIA catalog provider, or a local model provider is selectable without changing product code.
2. THE ShipCheck SHALL keep all model keys and API keys server-side.
3. THE ShipCheck SHALL NOT expose any model key or API key to browser code or other client code.
4. WHEN the ShipCheck_Advisor requires a model completion, THE ShipCheck SHALL call the ModelProvider from server-side code only.
5. THE ShipCheck SHALL run the deterministic fallback provider as the default provider on the public deployment so that the live demo requires no model key and incurs no model cost.
6. WHERE the HuggingFace_Provider is selected, THE ShipCheck SHALL call the Hugging Face Inference API from server-side code only using a server-side API key that is never exposed to client code.
7. WHERE the Bedrock_Provider is selected, THE ShipCheck SHALL invoke Amazon Bedrock from server-side Lambda code using the Lambda execution role's IAM permissions rather than a stored API key.
8. THE ShipCheck SHALL grant the Lambda execution role only the minimum Bedrock invoke permissions required.
9. THE ShipCheck SHALL treat Amazon Bedrock usage as consuming AWS credits or incurring per-token cost, because Bedrock has no free-inference tier, and THE ShipCheck SHALL retain the deterministic fallback provider as the always-free default so that the public demo incurs no model cost.

### Requirement 17: MCP External Context

**User Story:** As a launch owner, I want the advisor to enrich its recommendations with read-only external context, so that recommendations reflect real repository, issue, and documentation state without giving the model control over external systems.

#### Acceptance Criteria

1. WHERE external context is enabled, THE ShipCheck SHALL use MCP_Tools to fetch external context for the ShipCheck_Advisor, including get_public_repository_summary, get_open_issues, get_recent_pull_requests, get_release_notes, get_documentation_page, and inspect_public_github_reference.
2. THE ShipCheck SHALL constrain every MCP_Tool in the first release to read-only behavior.
3. IF an MCP_Tool would change GitHub data or ShipCheck data, THEN THE ShipCheck SHALL require explicit user confirmation before executing that MCP_Tool.
4. THE ShipCheck application SHALL manage the MCP connection and SHALL execute MCP_Tools, and THE model SHALL NOT manage the MCP connection or execute MCP_Tools directly.
5. WHEN the ShipCheck application invokes an MCP_Tool, THE ShipCheck application SHALL convert the MCP tool definitions to the model tool format and SHALL return the MCP_Tool results to the model.
6. THE ShipCheck SHALL limit external MCP fetches and external URL fetches in the first release to allowlisted public URLs and public repositories.
7. THE ShipCheck SHALL NOT send private repository content to external MCP fetches or external URL fetches by default.
8. THE ShipCheck SHALL use external context to enrich ShipCheck_Advisor recommendations only, and THE ShipCheck SHALL NOT provide external context to the Readiness_Engine.

### Requirement 18: Prompt-Injection and Untrusted Content Handling

**User Story:** As a security-conscious operator, I want all fetched external content treated as data and never as instructions, so that prompt-injection attempts cannot alter advisor behavior or leak secrets.

#### Acceptance Criteria

1. THE ShipCheck SHALL treat all Untrusted_External_Content, including repository files, issues, pull requests, documentation pages, and web pages, as data and SHALL NOT treat Untrusted_External_Content as instructions.
2. THE ShipCheck_Advisor system prompt SHALL instruct the model to disregard instructions embedded in Untrusted_External_Content.
3. THE ShipCheck_Advisor system prompt SHALL instruct the model to withhold secrets, system prompts, credentials, and hidden context from its output.
4. THE ShipCheck_Advisor SHALL use Untrusted_External_Content only as evidence for launch analysis.

### Requirement 19: Safe Data Flow and Privacy

**User Story:** As a launch owner, I want ShipCheck to minimize and protect the data it sends to the model and to keep me in control of changes, so that sensitive information stays private and recommendations remain reviewable.

#### Acceptance Criteria

1. WHEN a user submits an analysis request, THE ShipCheck SHALL load only the selected Launch, strip unnecessary personal information, prefer structured Launch data over whole documents, fetch only allowlisted public URLs, label external content as Untrusted_External_Content, request structured JSON from the model, validate the JSON against a schema, display recommendations with evidence links, and require user confirmation before saving any change.
2. THE ShipCheck SHALL NOT send to the ShipCheck_Advisor any API key, password, access token, private repository content by default, unnecessary name, email address, contact detail, customer record, or health, financial, or identity information.
3. THE ShipCheck SHALL indicate to the user that recommendations are AI-generated and require human review.
4. THE ShipCheck SHALL provide deletion of Launch data.
5. THE ShipCheck SHALL enforce a retention limit for AI request logs.
6. THE ShipCheck SHALL avoid storing raw prompts beyond what is required for operation.
7. THE ShipCheck SHALL serve requests over HTTPS.
8. THE ShipCheck SHALL keep all keys on the server.
9. THE ShipCheck SHALL apply rate limiting to analysis requests.
10. THE ShipCheck SHALL provide an abuse-reporting path.
11. THE ShipCheck SHALL process only content the user owns or is authorized to process, and THE ShipCheck SHALL support public repositories and example data only in the first release.

### Requirement 20: Compliance and Non-Decisioning Boundaries

**User Story:** As a product stakeholder, I want clear compliance and non-decisioning boundaries, so that ShipCheck stays within a safe scope and never presents itself as an approval authority.

#### Acceptance Criteria

1. THE ShipCheck SHALL NOT process health, legal, financial, biometric, or employment-decision data in the first release.
2. THE ShipCheck SHALL NOT make automated decisions about people.
3. THE ShipCheck SHALL present the Readiness_Assessment as decision support only, and THE ShipCheck SHALL NOT claim to provide legal, security, or compliance approval.
4. THE ShipCheck SHALL provide a privacy notice explaining what data is sent to the model provider.
5. THE ShipCheck SHALL record the exact model name, model provider, model license, and model access terms in the repository.

### Requirement 21: AI Provenance and Demo Disclaimer

**User Story:** As a product stakeholder, I want AI output labeled and the demonstration nature disclosed, so that visitors understand ShipCheck is a non-commercial demo and that AI recommendations are not a release approval.

#### Acceptance Criteria

1. THE ShipCheck SHALL label all ShipCheck_Advisor output as AI-generated and requiring human review.
2. THE ShipCheck SHALL display, on the public deployment, a disclaimer stating that ShipCheck is a non-commercial demonstration and that AI recommendations are not a release approval.
3. WHERE an opt-in model provider such as the HuggingFace_Provider, the Bedrock_Provider, or the NVIDIA catalog provider is used for a demonstration, THE ShipCheck SHALL disclose the specific model and provider used and SHALL note the applicable non-production, trial, or credit terms.
