// Domain types mirroring the Go backend's JSON (camelCase). These match the
// shapes returned by the ShipCheck API (internal/domain).

export type Category =
  | 'Requirements'
  | 'Engineering'
  | 'Quality Assurance'
  | 'Documentation'
  | 'Ownership'
  | 'Launch Operations'
  | 'Risks';

export const CATEGORIES: Category[] = [
  'Requirements',
  'Engineering',
  'Quality Assurance',
  'Documentation',
  'Ownership',
  'Launch Operations',
  'Risks',
];

export type CompletionState = 'complete' | 'in progress' | 'blocked' | 'incomplete';
export type Priority = 'Low' | 'Medium' | 'High';
export type Severity = 'Low' | 'Medium' | 'High';
export type Likelihood = 'Low' | 'Medium' | 'High';
export type RiskStatus = 'Open' | 'Mitigating' | 'Resolved' | 'Accepted';
export type ReadinessStatus =
  | 'Ready'
  | 'Conditionally Ready'
  | 'Not Ready'
  | 'Needs Review';
export type CategoryState = 'complete' | 'partial' | 'empty';

export interface Evidence {
  id: string;
  url?: string;
  note?: string;
  isPrivate: boolean;
}

export interface ChecklistItem {
  id: string;
  title: string;
  category: Category;
  owner?: string;
  priority: Priority;
  completionState: CompletionState;
  isCritical: boolean;
  evidence?: Evidence[];
}

export interface Risk {
  id: string;
  title: string;
  description?: string;
  severity: Severity;
  likelihood: Likelihood;
  owner?: string;
  mitigation?: string;
  status: RiskStatus;
  dueDate?: string;
  evidence?: Evidence[];
}

export interface ShareLink {
  token: string;
  enabled: boolean;
}

export interface LaunchBrief {
  whatIsReleasing?: string;
  audience?: string;
  successDefinition?: string;
  requirements?: string;
  constraints?: string;
  dependencies?: string;
}

export interface ReasonEntry {
  ruleId: string;
  message: string;
  itemId?: string;
  recommendedAction?: string;
}

export interface CategoryCompletion {
  category: Category;
  completeCount: number;
  totalCount: number;
  state: CategoryState;
}

export interface ReadinessAssessment {
  status: ReadinessStatus;
  reasons: ReasonEntry[];
  categories: CategoryCompletion[];
}

export interface Launch {
  id: string;
  name: string;
  description: string;
  targetDate: string;
  owner?: string;
  productArea?: string;
  repositoryUrl?: string;
  documentationUrl?: string;
  brief?: LaunchBrief;
  privateNotes?: string;
  checklistItems: ChecklistItem[];
  risks: Risk[];
  share?: ShareLink;
  assessment: ReadinessAssessment;
}

/** The id of the seeded demo launch (matches internal/seed.SeedLaunchID). */
export const SEED_LAUNCH_ID = 'team-inbox-2';

// ---- Advisor types (mirror internal/advisor) ----

export type AdvisorAction =
  | 'analyze'
  | 'find_blockers'
  | 'suggest_checklist'
  | 'review_risks'
  | 'prepare_review'
  | 'explain_readiness'
  | 'reanalyze';

export type RecommendationType =
  | 'missing_information'
  | 'likely_blocker'
  | 'possible_contradiction'
  | 'unowned_work'
  | 'suggested_checklist_item'
  | 'review_question';

export type EntityKind = 'checklistItem' | 'risk' | 'evidence' | 'launch';

export interface EvidenceRef {
  entityKind: EntityKind;
  entityId: string;
}

export interface Recommendation {
  type: RecommendationType;
  priority: Priority;
  title: string;
  reason: string;
  suggested_action: string;
  evidence: EvidenceRef[];
}

export interface LaunchAnalysis {
  recommendations: Recommendation[];
  suggested_questions: string[];
}

export interface AdvisorResult {
  analysis: LaunchAnalysis;
  discarded?: { index: number; reason: string }[];
  provider: string;
}

// ---- Public report (mirrors internal/domain.PublicReport) ----

export interface PublicItem {
  title: string;
  category: Category;
  completionState: CompletionState;
  critical: boolean;
}

export interface PublicRisk {
  title: string;
  severity: Severity;
  status: RiskStatus;
}

export interface PublicReport {
  name: string;
  targetDate: string;
  status: ReadinessStatus;
  categories: CategoryCompletion[];
  completedWork: PublicItem[];
  incompleteWork: PublicItem[];
  highPriorityRisks: PublicRisk[];
  openQuestions: string[];
  lastUpdated: string;
}

/** Field-level validation error returned by the API as HTTP 400. */
export interface ValidationError {
  field: string;
  message: string;
}
