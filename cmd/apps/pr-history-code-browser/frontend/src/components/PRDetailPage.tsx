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
                  <div style={{ marginBottom: '0.5rem', fontWeight: 500 }}>{entry.details}</div>
                  
                  {/* Show referenced commit */}
                  {entry.commit && (
                    <div style={{ 
                      marginBottom: '0.5rem', 
                      padding: '0.5rem', 
                      backgroundColor: '#f8f9fa',
                      borderLeft: '3px solid #3498db',
                      borderRadius: '3px'
                    }}>
                      <div style={{ fontSize: '0.9rem', marginBottom: '0.25rem' }}>
                        <strong>Commit:</strong>{' '}
                        <Link 
                          to={`/commits/${entry.commit.hash}`}
                          style={{ 
                            fontFamily: 'monospace', 
                            color: '#3498db',
                            textDecoration: 'none'
                          }}
                        >
                          {entry.commit.hash.substring(0, 8)}
                        </Link>
                      </div>
                      <div style={{ fontSize: '0.85rem', color: '#555' }}>
                        {entry.commit.subject}
                      </div>
                      <div style={{ fontSize: '0.8rem', color: '#7f8c8d', marginTop: '0.25rem' }}>
                        by {entry.commit.author_name} • {new Date(entry.commit.committed_at).toLocaleDateString()}
                      </div>
                    </div>
                  )}
                  
                  {/* Show referenced file */}
                  {entry.file && (
                    <div style={{ 
                      marginBottom: '0.5rem',
                      fontSize: '0.9rem',
                      fontFamily: 'monospace',
                      display: 'inline-block'
                    }}>
                      📄{' '}
                      {entry.file.id ? (
                        <Link 
                          to={`/files/${entry.file.id}`}
                          style={{
                            color: '#27ae60',
                            textDecoration: 'none',
                            backgroundColor: '#f0f0f0',
                            padding: '0.25rem 0.5rem',
                            borderRadius: '3px',
                          }}
                        >
                          {entry.file.path}
                        </Link>
                      ) : (
                        <span style={{
                          color: '#27ae60',
                          backgroundColor: '#f0f0f0',
                          padding: '0.25rem 0.5rem',
                          borderRadius: '3px',
                        }}>
                          {entry.file.path}
                        </span>
                      )}
                    </div>
                  )}
                  
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
                
                {/* Show referenced commit */}
                {note.commit && (
                  <div style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                    <strong>Related to commit:</strong>{' '}
                    <Link 
                      to={`/commits/${note.commit.hash}`}
                      style={{ 
                        fontFamily: 'monospace', 
                        color: '#3498db',
                        textDecoration: 'none'
                      }}
                    >
                      {note.commit.hash.substring(0, 8)}
                    </Link>
                    {' - '}{note.commit.subject}
                  </div>
                )}
                
                {/* Show referenced file */}
                {note.file && (
                  <div style={{ 
                    marginTop: '0.5rem',
                    fontSize: '0.9rem',
                  }}>
                    <strong>Related to file:</strong>{' '}
                    {note.file.id ? (
                      <Link 
                        to={`/files/${note.file.id}`}
                        style={{
                          fontFamily: 'monospace',
                          color: '#27ae60',
                          textDecoration: 'none'
                        }}
                      >
                        {note.file.path}
                      </Link>
                    ) : (
                      <span style={{ fontFamily: 'monospace', color: '#27ae60' }}>
                        {note.file.path}
                      </span>
                    )}
                  </div>
                )}
                
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

