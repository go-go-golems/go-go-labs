import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { File } from '../types';

export function FilesPage() {
  const [files, setFiles] = useState<File[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [prefix, setPrefix] = useState('');
  const [offset, setOffset] = useState(0);
  const limit = 100;

  useEffect(() => {
    setLoading(true);
    api
      .getFiles(limit, offset, prefix)
      .then(setFiles)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [offset, prefix]);

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
    return <div className="loading">Loading files...</div>;
  }

  if (error) {
    return <div className="error">Error loading files: {error}</div>;
  }

  return (
    <div>
      <div className="page-header">
        <h2>Files</h2>
        <p>Browse files tracked in the repository</p>
      </div>

      <div className="search-box">
        <form onSubmit={handleSearch}>
          <input
            type="text"
            className="search-input"
            placeholder="Filter by path prefix..."
            value={prefix}
            onChange={(e) => setPrefix(e.target.value)}
          />
        </form>
      </div>

      <ul className="file-list">
        {files.map((file) => (
          <li key={file.id} className="file-item">
            <code className="file-path">{file.path}</code>
          </li>
        ))}
      </ul>

      {files.length === 0 && (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#7f8c8d' }}>
          No files found
        </div>
      )}

      <div className="pagination">
        <button onClick={handlePrevious} disabled={offset === 0}>
          Previous
        </button>
        <span>
          Showing {offset + 1} - {offset + files.length}
        </span>
        <button onClick={handleNext} disabled={files.length < limit}>
          Next
        </button>
      </div>
    </div>
  );
}

