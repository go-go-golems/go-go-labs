import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { PR } from '../types';

export function PRsPage() {
  const navigate = useNavigate();
  const [prs, setPRs] = useState<PR[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getPRs()
      .then(setPRs)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="loading">Loading PRs...</div>;
  }

  if (error) {
    return <div className="error">Error loading PRs: {error}</div>;
  }

  return (
    <div>
      <div className="page-header">
        <h2>Pull Requests</h2>
        <p>Browse PR slices and their work tracking</p>
      </div>

      <ul className="pr-list">
        {prs.map((pr) => (
          <li key={pr.id} className="pr-item" onClick={() => navigate(`/prs/${pr.id}`)}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <h3 style={{ fontSize: '1.25rem', marginBottom: '0.5rem' }}>{pr.name}</h3>
                <p style={{ color: '#7f8c8d', marginBottom: '0.5rem' }}>{pr.description}</p>
                <div style={{ fontSize: '0.9rem', color: '#7f8c8d' }}>
                  Created: {new Date(pr.created_at).toLocaleDateString()}
                  {pr.updated_at && ` • Updated: ${new Date(pr.updated_at).toLocaleDateString()}`}
                </div>
              </div>
              <span className={`pr-status ${pr.status}`}>{pr.status}</span>
            </div>
          </li>
        ))}
      </ul>

      {prs.length === 0 && (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#7f8c8d' }}>
          No PRs found
        </div>
      )}
    </div>
  );
}

