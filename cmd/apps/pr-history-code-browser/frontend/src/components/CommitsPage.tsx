import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { Commit } from '../types';

export function CommitsPage() {
  const navigate = useNavigate();
  const [commits, setCommits] = useState<Commit[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [offset, setOffset] = useState(0);
  const limit = 50;

  useEffect(() => {
    setLoading(true);
    api
      .getCommits(limit, offset, search)
      .then(setCommits)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [offset, search]);

  const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setOffset(0);
  };

  const handleNext = () => {
    setOffset(offset + limit);
  };

  const handlePrevious = () => {
    setOffset(Math.max(0, offset - limit));
  };

  if (loading) {
    return <div className="loading">Loading commits...</div>;
  }

  if (error) {
    return <div className="error">Error loading commits: {error}</div>;
  }

  return (
    <div>
      <div className="page-header">
        <h2>Commits</h2>
        <p>Browse the repository commit history</p>
      </div>

      <div className="search-box">
        <form onSubmit={handleSearch}>
          <input
            type="text"
            className="search-input"
            placeholder="Search commits (hash, subject, or body)..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </form>
      </div>

      <ul className="commit-list">
        {commits.map((commit) => (
          <li
            key={commit.id}
            className="commit-item"
            onClick={() => navigate(`/commits/${commit.hash}`)}
          >
            <div>
              <span className="commit-hash">{commit.hash.substring(0, 8)}</span>
            </div>
            <div className="commit-subject">{commit.subject}</div>
            <div className="commit-meta">
              <span>👤 {commit.author_name}</span>
              <span>📅 {new Date(commit.committed_at).toLocaleString()}</span>
            </div>
          </li>
        ))}
      </ul>

      {commits.length === 0 && (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#7f8c8d' }}>
          No commits found
        </div>
      )}

      <div className="pagination">
        <button onClick={handlePrevious} disabled={offset === 0}>
          Previous
        </button>
        <span>
          Showing {offset + 1} - {offset + commits.length}
        </span>
        <button onClick={handleNext} disabled={commits.length < limit}>
          Next
        </button>
      </div>
    </div>
  );
}

