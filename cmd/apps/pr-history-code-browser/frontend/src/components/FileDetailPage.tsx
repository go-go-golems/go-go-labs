import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api/client';
import type { FileWithHistory } from '../types';

export function FileDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [fileDetails, setFileDetails] = useState<FileWithHistory | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;

    api
      .getFileDetails(parseInt(id, 10))
      .then(setFileDetails)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return <div className="loading">Loading file details...</div>;
  }

  if (error) {
    return <div className="error">Error loading file: {error}</div>;
  }

  if (!fileDetails) {
    return null;
  }

  return (
    <div>
      <Link to="/files" className="back-link">
        ← Back to Files
      </Link>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">File Details</h3>
        </div>
        <div>
          <div style={{ 
            fontFamily: 'monospace', 
            fontSize: '1.1rem', 
            marginBottom: '1rem',
            padding: '0.5rem',
            backgroundColor: '#f0f0f0',
            borderRadius: '3px'
          }}>
            📄 {fileDetails.path}
          </div>
          <div style={{ color: '#7f8c8d', fontSize: '0.9rem' }}>
            Total commits: {fileDetails.commit_count}
          </div>
        </div>
      </div>

      {fileDetails.recent_commits && fileDetails.recent_commits.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Recent Commits ({fileDetails.recent_commits.length})</h3>
          </div>
          <div>
            {fileDetails.recent_commits.map((commit) => (
              <div
                key={commit.id}
                style={{
                  padding: '1rem',
                  borderBottom: '1px solid #ecf0f1',
                }}
              >
                <div style={{ marginBottom: '0.5rem' }}>
                  <Link
                    to={`/commits/${commit.hash}`}
                    style={{
                      fontFamily: 'monospace',
                      color: '#3498db',
                      textDecoration: 'none',
                      fontWeight: 500,
                    }}
                  >
                    {commit.hash.substring(0, 8)}
                  </Link>
                  <span style={{ marginLeft: '1rem', fontSize: '1rem' }}>
                    {commit.subject}
                  </span>
                </div>
                <div style={{ fontSize: '0.85rem', color: '#7f8c8d' }}>
                  {commit.author_name} • {new Date(commit.committed_at).toLocaleDateString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {fileDetails.pr_references && fileDetails.pr_references.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Referenced in PRs ({fileDetails.pr_references.length})</h3>
          </div>
          <div>
            <p style={{ padding: '0.5rem 1rem', fontSize: '0.9rem', color: '#7f8c8d' }}>
              This file was referenced in the following PRs:
            </p>
            {fileDetails.pr_references.map((prRef, idx) => (
              <div
                key={idx}
                style={{
                  padding: '1rem',
                  borderLeft: '3px solid #3498db',
                  marginBottom: '0.5rem',
                  backgroundColor: '#f8f9fa',
                  cursor: 'pointer',
                }}
                onClick={() => (window.location.href = `/prs/${prRef.pr_id}`)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                  <Link 
                    to={`/prs/${prRef.pr_id}`}
                    style={{ 
                      fontWeight: 500, 
                      fontSize: '1rem',
                      color: '#3498db',
                      textDecoration: 'none'
                    }}
                  >
                    {prRef.pr_name}
                  </Link>
                  <span className="action-badge">{prRef.action}</span>
                </div>
                <div style={{ fontSize: '0.9rem', marginBottom: '0.25rem' }}>
                  {prRef.details}
                </div>
                <div style={{ fontSize: '0.85rem', color: '#7f8c8d' }}>
                  {new Date(prRef.created_at).toLocaleDateString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {fileDetails.related_files && fileDetails.related_files.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Files Often Changed Together ({fileDetails.related_files.length})</h3>
          </div>
          <div>
            <p style={{ padding: '0.5rem 1rem', fontSize: '0.9rem', color: '#7f8c8d' }}>
              These files were frequently modified in the same commits as this file.
            </p>
            {fileDetails.related_files.map((relatedFile, idx) => (
              <div
                key={idx}
                style={{
                  padding: '0.75rem 1rem',
                  borderBottom: '1px solid #ecf0f1',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <code style={{ 
                  fontFamily: 'monospace', 
                  fontSize: '0.9rem',
                  color: '#27ae60'
                }}>
                  {relatedFile.path}
                </code>
                <span style={{ 
                  fontSize: '0.85rem', 
                  color: '#7f8c8d',
                  backgroundColor: '#f8f9fa',
                  padding: '0.25rem 0.5rem',
                  borderRadius: '3px'
                }}>
                  {relatedFile.change_count} co-changes
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {fileDetails.notes && fileDetails.notes.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Analysis Notes ({fileDetails.notes.length})</h3>
          </div>
          <div>
            {fileDetails.notes.map((note) => (
              <div key={note.id} className="note-item">
                <div className="note-type">{note.note_type}</div>
                <div className="note-text">{note.note}</div>
                
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

