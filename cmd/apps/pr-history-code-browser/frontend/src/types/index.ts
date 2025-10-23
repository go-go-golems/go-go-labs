export interface Commit {
  id: number;
  hash: string;
  parents: string;
  author_name: string;
  author_email: string;
  authored_at: string;
  committer_name: string;
  committer_email: string;
  committed_at: string;
  subject: string;
  body: string;
  document_summary: any;
}

export interface FileChange {
  file_id: number;
  path: string;
  change_type: string;
  old_path?: string;
  additions: number;
  deletions: number;
}

export interface CommitSymbol {
  commit_id: number;
  file_id: number;
  symbol_name: string;
  symbol_kind: string;
}

export interface CommitDetails {
  commit: Commit;
  files: FileChange[];
  symbols: CommitSymbol[];
}

export interface File {
  id: number;
  path: string;
}

export interface PR {
  id: number;
  name: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface PRChangelog {
  id: number;
  pr_id?: number;
  commit_id?: number;
  file_id?: number;
  action: string;
  details: string;
  created_at: string;
}

export interface AnalysisNote {
  id: number;
  commit_id?: number;
  file_id?: number;
  note_type: string;
  note: string;
  tags: string;
  created_at: string;
}

export interface PRDetails {
  id: number;
  name: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
  changelog: PRChangelog[];
  notes: AnalysisNote[];
}

export interface Stats {
  commit_count: number;
  file_count: number;
  pr_count: number;
  analysis_note_count: number;
  earliest_commit: string;
  latest_commit: string;
  pr_status_counts: Record<string, number>;
}

