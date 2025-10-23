import { Link, Outlet, useLocation } from 'react-router-dom';

export function Layout() {
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path ? 'active' : '';
  };

  return (
    <div className="app">
      <nav className="nav">
        <div className="nav-content">
          <h1>PR History & Code Browser</h1>
          <ul className="nav-links">
            <li>
              <Link to="/" className={isActive('/')}>
                Home
              </Link>
            </li>
            <li>
              <Link to="/commits" className={isActive('/commits')}>
                Commits
              </Link>
            </li>
            <li>
              <Link to="/prs" className={isActive('/prs')}>
                PRs
              </Link>
            </li>
            <li>
              <Link to="/files" className={isActive('/files')}>
                Files
              </Link>
            </li>
            <li>
              <Link to="/notes" className={isActive('/notes')}>
                Notes
              </Link>
            </li>
          </ul>
        </div>
      </nav>
      <main className="container">
        <Outlet />
      </main>
    </div>
  );
}

