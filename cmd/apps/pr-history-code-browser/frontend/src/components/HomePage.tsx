import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { Stats } from '../types';

export function HomePage() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getStats()
      .then(setStats)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="loading">Loading statistics...</div>;
  }

  if (error) {
    return <div className="error">Error loading statistics: {error}</div>;
  }

  if (!stats) {
    return null;
  }

  return (
    <div>
      <div className="page-header">
        <h2>Repository Overview</h2>
        <p>Browse git history, PRs, and analysis notes from the database</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card" style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}>
          <div className="stat-value">{stats.commit_count}</div>
          <div className="stat-label">Total Commits</div>
        </div>

        <div className="stat-card" style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
          <div className="stat-value">{stats.file_count}</div>
          <div className="stat-label">Files Tracked</div>
        </div>

        <div className="stat-card" style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
          <div className="stat-value">{stats.pr_count}</div>
          <div className="stat-label">Pull Requests</div>
        </div>

        <div className="stat-card" style={{ background: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' }}>
          <div className="stat-value">{stats.analysis_note_count}</div>
          <div className="stat-label">Analysis Notes</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Repository Timeline</h3>
        </div>
        <div>
          <p>
            <strong>Earliest Commit:</strong> {new Date(stats.earliest_commit).toLocaleString()}
          </p>
          <p>
            <strong>Latest Commit:</strong> {new Date(stats.latest_commit).toLocaleString()}
          </p>
        </div>
      </div>

      {stats.pr_status_counts && Object.keys(stats.pr_status_counts).length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">PR Status Breakdown</h3>
          </div>
          <div>
            {Object.entries(stats.pr_status_counts).map(([status, count]) => (
              <div key={status} style={{ marginBottom: '0.5rem' }}>
                <span className={`pr-status ${status}`}>{status}</span>
                <span style={{ marginLeft: '1rem', color: '#7f8c8d' }}>{count} PR(s)</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Quick Links</h3>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <Link to="/commits" style={{ color: '#3498db', textDecoration: 'none', fontWeight: 500 }}>
            Browse Commits →
          </Link>
          <Link to="/prs" style={{ color: '#3498db', textDecoration: 'none', fontWeight: 500 }}>
            View PRs →
          </Link>
          <Link to="/files" style={{ color: '#3498db', textDecoration: 'none', fontWeight: 500 }}>
            Explore Files →
          </Link>
          <Link to="/notes" style={{ color: '#3498db', textDecoration: 'none', fontWeight: 500 }}>
            Read Notes →
          </Link>
        </div>
      </div>
    </div>
  );
}

