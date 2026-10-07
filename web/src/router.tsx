import { createBrowserRouter, Navigate } from 'react-router'
import { AppShell } from './routes/AppShell'
import { HomePage } from './routes/HomePage'
import { LoginCallbackPage } from './routes/LoginCallbackPage'
import { LoginPage } from './routes/LoginPage'
import { RequireSession } from './routes/RequireSession'
import { RequireSetup } from './routes/RequireSetup'
import { SettingsPage } from './routes/SettingsPage'
import { SetupPage } from './routes/SetupPage'

export function createAppRouter() {
  return createBrowserRouter([
    { path: '/login', element: <LoginPage /> },
    { path: '/login/callback', element: <LoginCallbackPage /> },
    {
      element: <RequireSession />,
      children: [
        { path: '/setup', element: <SetupPage /> },
        {
          element: <RequireSetup />,
          children: [
            {
              element: <AppShell />,
              children: [
                { index: true, element: <HomePage /> },
                { path: '/settings', element: <SettingsPage /> },
              ],
            },
          ],
        },
      ],
    },
    { path: '*', element: <Navigate to="/" replace /> },
  ])
}

