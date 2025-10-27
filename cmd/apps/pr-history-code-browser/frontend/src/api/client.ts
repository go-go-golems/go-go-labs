import type { Commit, CommitDetails, PR, PRDetails, File, FileWithHistory, AnalysisNote, Stats } from '../types';

const API_BASE = '/api';

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`API request failed: ${response.statusText}`);
  }
  return response.json();
}

export const api = {
  // Statistics
  async getStats(): Promise<Stats> {
    return fetchJSON<Stats>(`${API_BASE}/stats`);
  },

  // Commits
  async getCommits(limit: number = 50, offset: number = 0, search?: string): Promise<Commit[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    if (search) {
      params.append('search', search);
    }
    return fetchJSON<Commit[]>(`${API_BASE}/commits?${params}`);
  },

  async getCommit(hash: string): Promise<CommitDetails> {
    return fetchJSON<CommitDetails>(`${API_BASE}/commits/${hash}`);
  },

  // PRs
  async getPRs(): Promise<PR[]> {
    return fetchJSON<PR[]>(`${API_BASE}/prs`);
  },

  async getPR(id: number): Promise<PRDetails> {
    return fetchJSON<PRDetails>(`${API_BASE}/prs/${id}`);
  },

  // Files
  async getFiles(limit: number = 100, offset: number = 0, prefix?: string): Promise<File[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    if (prefix) {
      params.append('prefix', prefix);
    }
    return fetchJSON<File[]>(`${API_BASE}/files?${params}`);
  },

  async getFileHistory(fileId: number, limit: number = 50): Promise<Commit[]> {
    return fetchJSON<Commit[]>(`${API_BASE}/files/${fileId}/history?limit=${limit}`);
  },

  async getFileDetails(fileId: number): Promise<FileWithHistory> {
    return fetchJSON<FileWithHistory>(`${API_BASE}/files/${fileId}/details`);
  },

  // Analysis Notes
  async getAnalysisNotes(
    limit: number = 50,
    offset: number = 0,
    noteType?: string,
    tags?: string
  ): Promise<AnalysisNote[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    if (noteType) {
      params.append('type', noteType);
    }
    if (tags) {
      params.append('tags', tags);
    }
    return fetchJSON<AnalysisNote[]>(`${API_BASE}/notes?${params}`);
  },
};

