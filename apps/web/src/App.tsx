import { createBrowserRouter, Outlet } from 'react-router-dom'
import { Dashboard } from '@/pages/Dashboard'
import { NotFound } from '@/pages/NotFound'

function RootLayout() {
  return (
    <div>
      <Outlet />
    </div>
  )
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: '*', element: <NotFound /> },
    ],
  },
])
