import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { HomePage } from './components/HomePage';
import { CommitsPage } from './components/CommitsPage';
import { CommitDetailPage } from './components/CommitDetailPage';
import { PRsPage } from './components/PRsPage';
import { PRDetailPage } from './components/PRDetailPage';
import { FilesPage } from './components/FilesPage';
import { NotesPage } from './components/NotesPage';
import './App.css';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="commits" element={<CommitsPage />} />
          <Route path="commits/:hash" element={<CommitDetailPage />} />
          <Route path="prs" element={<PRsPage />} />
          <Route path="prs/:id" element={<PRDetailPage />} />
          <Route path="files" element={<FilesPage />} />
          <Route path="notes" element={<NotesPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

