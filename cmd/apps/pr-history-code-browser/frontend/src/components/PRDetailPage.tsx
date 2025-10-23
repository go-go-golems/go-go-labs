import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api/client';
import type { PRDetails } from '../types';

export function PRDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [pr, setPR] = useState<PRDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;

    api
      .getPR(parseInt(id, 10))
      .then(setPR)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return <div className="loading">Loading PR details...</div>;
  }

  if (error) {
    return <div className="error">Error loading PR: {error}</div>;
  }

  if (!pr) {
    return null;
  }

  return (
    <div>
      <Link to="/prs" className="back-link">
        ← Back to PRs
      </Link>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">{pr.name}</h3>
          <span className={`pr-status ${pr.status}`}>{pr.status}</span>
        </div>
        <div>
          <p style={{ fontSize: '1.1rem', marginBottom: '1rem' }}>{pr.description}</p>
          <div style={{ color: '#7f8c8d', fontSize: '0.9rem' }}>
            <div>Created: {new Date(pr.created_at).toLocaleDateString()}</div>
            {pr.updated_at && <div>Updated: {new Date(pr.updated_at).toLocaleDateString()}</div>}
          </div>
        </div>
      </div>

      {pr.changelog && pr.changelog.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Changelog ({pr.changelog.length})</h3>
          </div>
          <div>
            {pr.changelog.map((entry) => (
              <div key={entry.id} className="changelog-item">
                <span className="action-badge">{entry.action}</span>
                <div style={{ flex: 1 }}>
                  <div style={{ marginBottom: '0.25rem' }}>{entry.details}</div>
                  <div style={{ fontSize: '0.85rem', color: '#7f8c8d' }}>
                    {new Date(entry.created_at).toLocaleString()}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {pr.notes && pr.notes.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Analysis Notes ({pr.notes.length})</h3>
          </div>
          <div>
            {pr.notes.map((note) => (
              <div key={note.id} className="note-item">
                <div className="note-type">{note.note_type}</div>
                <div className="note-text">{note.note}</div>
                {note.tags && (
                  <div className="note-tags">
                    Tags: {note.tags.split(',').map((tag) => (
                      <span key={tag.trim()} className="tag">
                        {tag.trim()}
                      </span>
                    ))}
                  </div>
                )}
                <div style={{ marginTop: '0.5rem', fontSize: '0.85rem', color: '#7f8c8d' }}>
                  {new Date(note.created_at).toLocaleString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

