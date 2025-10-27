import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api/client';
import type { CommitDetails } from '../types';

export function CommitDetailPage() {
  const { hash } = useParams<{ hash: string }>();
  const [details, setDetails] = useState<CommitDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!hash) return;

    api
      .getCommit(hash)
      .then(setDetails)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [hash]);

  if (loading) {
    return <div className="loading">Loading commit details...</div>;
  }

  if (error) {
    return <div className="error">Error loading commit: {error}</div>;
  }

  if (!details) {
    return null;
  }

  const { commit, files, symbols, pr_associations, notes } = details;

  return (
    <div>
      <Link to="/commits" className="back-link">
        ← Back to commits
      </Link>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Commit Details</h3>
          <span className="commit-hash">{commit.hash}</span>
        </div>
        <div>
          <h4 style={{ marginBottom: '1rem', fontSize: '1.25rem' }}>{commit.subject}</h4>
          {commit.body && (
            <pre
              style={{
                whiteSpace: 'pre-wrap',
                backgroundColor: '#f8f9fa',
                padding: '1rem',
                borderRadius: '4px',
                marginBottom: '1rem',
              }}
            >
              {commit.body}
            </pre>
          )}
          <div className="commit-meta">
            <span>👤 Author: {commit.author_name} ({commit.author_email})</span>
            <span>📅 {new Date(commit.committed_at).toLocaleString()}</span>
          </div>
          {commit.parents && (
            <div style={{ marginTop: '0.5rem', color: '#7f8c8d' }}>
              <strong>Parents:</strong> {commit.parents}
            </div>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Changed Files ({files.length})</h3>
        </div>
        <div>
          {files.map((file, idx) => (
            <div key={idx} className="file-change">
              <span className={`change-type ${file.change_type}`}>{file.change_type}</span>
              {file.file_id ? (
                <Link 
                  to={`/files/${file.file_id}`}
                  style={{
                    textDecoration: 'none',
                    color: 'inherit',
                  }}
                >
                  <code className="file-path" style={{ cursor: 'pointer' }}>{file.path}</code>
                </Link>
              ) : (
                <code className="file-path">{file.path}</code>
              )}
              {file.old_path && file.old_path !== file.path && (
                <span style={{ color: '#7f8c8d', fontSize: '0.9rem' }}>
                  (from {file.old_path})
                </span>
              )}
              <span className="stats">
                <span style={{ color: '#2ecc71' }}>+{file.additions}</span>
                <span style={{ color: '#e74c3c' }}>-{file.deletions}</span>
              </span>
            </div>
          ))}
        </div>
      </div>

      {pr_associations && pr_associations.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Related PRs ({pr_associations.length})</h3>
          </div>
          <div>
            {pr_associations.map((assoc, idx) => (
              <div
                key={idx}
                style={{
                  padding: '1rem',
                  borderLeft: '3px solid #3498db',
                  marginBottom: '1rem',
                  backgroundColor: '#f8f9fa',
                  cursor: 'pointer',
                }}
                onClick={() => (window.location.href = `/prs/${assoc.pr_id}`)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <strong style={{ fontSize: '1.1rem' }}>{assoc.pr_name}</strong>
                  <span className="action-badge">{assoc.action}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {notes && notes.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Analysis Notes ({notes.length})</h3>
          </div>
          <div>
            {notes.map((note) => (
              <div key={note.id} className="note-item">
                <div className="note-type">{note.note_type}</div>
                <div className="note-text">{note.note}</div>
                {note.tags && (
                  <div className="note-tags">
                    Tags:{' '}
                    {note.tags.split(',').map((tag) => (
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

      {symbols.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Symbols ({symbols.length})</h3>
          </div>
          <div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #ecf0f1', textAlign: 'left' }}>
                  <th style={{ padding: '0.5rem' }}>Symbol Name</th>
                  <th style={{ padding: '0.5rem' }}>Kind</th>
                </tr>
              </thead>
              <tbody>
                {symbols.map((symbol, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid #ecf0f1' }}>
                    <td style={{ padding: '0.5rem' }}>
                      <code>{symbol.symbol_name}</code>
                    </td>
                    <td style={{ padding: '0.5rem' }}>
                      <span className="tag">{symbol.symbol_kind}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

