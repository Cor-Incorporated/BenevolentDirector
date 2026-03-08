import { Link } from 'react-router-dom'

export function NotFound() {
  return (
    <main>
      <h1>404 - Page Not Found</h1>
      <p>The page you are looking for does not exist.</p>
      <Link to="/">Back to Dashboard</Link>
    </main>
  )
}
