import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { AnalysisNote } from '../types';

export function NotesPage() {
  const [notes, setNotes] = useState<AnalysisNote[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tags, setTags] = useState('');
  const [offset, setOffset] = useState(0);
  const limit = 50;

  useEffect(() => {
    setLoading(true);
    api
      .getAnalysisNotes(limit, offset, undefined, tags)
      .then(setNotes)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [offset, tags]);

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
    return <div className="loading">Loading analysis notes...</div>;
  }

  if (error) {
    return <div className="error">Error loading notes: {error}</div>;
  }

  return (
    <div>
      <div className="page-header">
        <h2>Analysis Notes</h2>
        <p>Browse manual analysis notes and annotations</p>
      </div>

      <div className="search-box">
        <form onSubmit={handleSearch}>
          <input
            type="text"
            className="search-input"
            placeholder="Filter by tags..."
            value={tags}
            onChange={(e) => setTags(e.target.value)}
          />
        </form>
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

      {notes.length === 0 && (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#7f8c8d' }}>
          No notes found
        </div>
      )}

      <div className="pagination">
        <button onClick={handlePrevious} disabled={offset === 0}>
          Previous
        </button>
        <span>
          Showing {offset + 1} - {offset + notes.length}
        </span>
        <button onClick={handleNext} disabled={notes.length < limit}>
          Next
        </button>
      </div>
    </div>
  );
}

